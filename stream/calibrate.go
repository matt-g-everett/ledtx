package stream


import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	binSimilarityDistance float64 = 3.0
	binHitThreshold int32 = 100
	iterations int = 30
)

type Calibrate struct {
	config Config
	client mqtt.Client
	C chan bool
	started bool

	iteration int
	bins map[Point]map[int]int
	frame *Frame
	msg chan CalibrationMessage
	rawData []*RawCalibrationData
	isEven bool
	aggregated *AggregatedData

	binWriteLock sync.Mutex
}

func NewCalibrate(config Config, client mqtt.Client) *Calibrate {
	c := new(Calibrate)
	c.config = config
	c.client = client
	c.C = make(chan bool)
	c.started = false
	return c
}

func (c *Calibrate) prepareFS() {
	os.RemoveAll("caldata")
	os.MkdirAll("caldata/raw", 0755)
	os.MkdirAll("caldata/pixels", 0755)
}

func (c *Calibrate) prepareFrame(frame *Frame, litLength int, isEven bool) []int32 {
	remainder := 1
	if isEven {
		remainder = 0
	}

	pixelCount := len(frame.pixels)
	lit := make([]int32, pixelCount, pixelCount)
	for i := 0; i < pixelCount; i++ {
		litlen := 1 << litLength
		if (int(math.Floor(float64(i) / float64(litlen))) % 2) == remainder {
			frame.pixels[i], _ = colorful.Hex("#202020")
			lit[i] = 1
		} else {
			frame.pixels[i], _ = colorful.Hex("#000000")
			lit[i] = -1
		}
	}

	return lit
}

func (c *Calibrate) handleClientMessages(client mqtt.Client, msg mqtt.Message) {
	var message CalibrationMessage
	json.Unmarshal(msg.Payload(), &message)

	if !c.started && message.Type == "start" {
		go c.runCalibration()
	} else if c.started && message.Type == "data" {
		c.msg<- message
	}
}

func (c *Calibrate) runCalibration() {
	c.started = true
	c.frame = NewFrame()
	pixelCount := len(c.frame.pixels)
	litLength := int(math.Ceil(math.Log2(float64(pixelCount)))) - 1
	c.aggregated = &AggregatedData { Bins: make([]*Bin, 0, 5000) }
	c.msg = make(chan CalibrationMessage)
	rawDataCount := 2 * (litLength + 1) * iterations
	c.rawData = make([]*RawCalibrationData, 0, rawDataCount)

	c.prepareFrame(c.frame, litLength, c.isEven) // Do this early allow the camera to adjust exposure
	c.C<- true
	time.Sleep(2 * time.Second)

	c.prepareFS()
	c.isEven = true
	patternNumber := 0
	var importWaitGroup sync.WaitGroup
	for ; litLength >= 0 && c.started; {
		lit := c.prepareFrame(c.frame, litLength, c.isEven)
		time.Sleep(200 * time.Millisecond)

		for i := 0; i < iterations; i++ {
			token := c.client.Publish(c.config.Mqtt.Topics.CalibrateServer, 0, false, "snapshot")
			token.Wait()
			log.Println("Published snapshot")

			t := time.NewTimer(time.Second)
			select {
			case msg := <-c.msg:
				log.Println("Received snapshot")
				importWaitGroup.Add(1)
				go c.importCalibrationMessage(msg, lit, patternNumber, i, &importWaitGroup)
			case <-t.C:
				log.Println("Message timed-out, retrying")
			}
		}

		patternNumber++

		if !c.isEven {
			litLength--
		}
		c.isEven = !c.isEven
	}

	importWaitGroup.Wait()
	log.Println("########## DONE CAPTURING")

	c.aggregate()
	log.Printf("Bin count (total): %d", len(c.aggregated.Bins))

	highHitData := &AggregatedData { Bins: make([]*Bin, 0, 10000) }
	for _, b := range c.aggregated.Bins {
		if b.Hits >= binHitThreshold {
			highHitData.Bins = append(highHitData.Bins, b)
		}
	}
	log.Printf("Bin count (hits): %d", len(highHitData.Bins))

	c.storeAggregatedData(c.aggregated)
	c.storeAggregatedData(highHitData)

	// log.Printf("Count: %d", len(c.bins))
	// report := make(map[int]Point)
	// reportCount := make(map[int]int)


	// for i := 0; i < pixelCount; i++ {
	// 	highestCount := 0
	// 	highestBin := Point{0, 0}
	// 	for k, v := range c.bins {
	// 		if v[i] > highestCount {
	// 			highestBin = k
	// 			highestCount = v[i]
	// 		}
	// 	}

	// 	report[i] = highestBin
	// 	reportCount[i] = highestCount
	// }

	// for i := 0; i < pixelCount; i++ {
	// 	loc, found := report[i]
	// 	count, _ := reportCount[i]
	// 	if found {
	// 		log.Printf("%d: %v (%d)", i, loc, count)
	// 	} else {
	// 		log.Printf("%d:", i)
	// 	}
	// }
}

func (c *Calibrate) aggregate() {
	for _, r := range c.rawData {
		for _, l := range r.Locations {
			c.incrementBin(l, r.Pixels)
		}
	}

	c.storeAggregatedData(c.aggregated)
}

func (c *Calibrate) doIncrementBin(bin *Bin, lit []int32) {
	atomic.AddInt32(&bin.Hits, 1)
	for j := 0; j < len(lit); j++ {
		atomic.AddInt32(&bin.Pixels[j], lit[j])
	}
}

func (c *Calibrate) incrementBin(binLocation Point, lit []int32) {
	found := false
	for i := 0; i < len(c.aggregated.Bins); i++ {
		if isBin(binLocation, c.aggregated.Bins[i].Location, binSimilarityDistance) {
			c.doIncrementBin(c.aggregated.Bins[i], lit)
			found = true
			break
		}
	}

	if !found {
		c.binWriteLock.Lock()
		// Search again, another goroutine may have beaten us to it
		found = false
		var foundBin *Bin
		for i := 0; i < len(c.aggregated.Bins); i++ {
			if isBin(binLocation, c.aggregated.Bins[i].Location, binSimilarityDistance) {
				found = true
				foundBin = c.aggregated.Bins[i]
				break
			}
		}

		// If the bin is still not present while we're locked, add it
		if !found {
			litCopy := make([]int32, len(lit))
			copy(litCopy, lit)
			c.aggregated.Bins = append(c.aggregated.Bins, &Bin { Location: binLocation, Pixels: litCopy, Hits: 1 })
		}
		c.binWriteLock.Unlock()

		// If we found it then another goroutine added it, so just increment as usual
		if found {
			c.doIncrementBin(foundBin, lit)
		}
	}
}

func isBin(a Point, b Point, threshold float64) bool {
	return math.Sqrt(math.Pow(math.Abs(a.X - b.X), 2.0) + math.Pow(math.Abs(a.Y - b.Y), 2.0)) < threshold
}

func (c *Calibrate) convertCalibrationMessage(msg CalibrationMessage, lit []int32) *RawCalibrationData {
	pointCount := len(msg.Locations) / 2
	r := &RawCalibrationData {
		Pixels: lit,
		Locations: make([]Point, pointCount, pointCount),
	}

	for i := 0; i < pointCount; i++ {
		r.Locations[i] = Point {
			X: msg.Locations[i * 2],
			Y: msg.Locations[(i * 2) + 1],
		}
	}

	return r
}

func (c *Calibrate) storeRawData(data *RawCalibrationData, patternNumber int, capture int) {
	filePath := fmt.Sprintf("caldata/raw/raw-p%03d-%03d.json", patternNumber, capture)
	f, err := os.OpenFile(filePath, os.O_CREATE | os.O_WRONLY, 0664)
	if err != nil {
		panic(err.Error)
	}

	serialised, err := json.Marshal(data)
	if err != nil {
		f.Close()
		panic(err.Error)
	}
	_, err = f.Write(serialised)
	if err != nil {
		f.Close()
		panic(err.Error)
	}
	f.Close()
}

func (c *Calibrate) storeAggregatedData(data *AggregatedData) {
	filePath := fmt.Sprintf("caldata/aggregated.json")
	f, err := os.OpenFile(filePath, os.O_CREATE | os.O_WRONLY, 0664)
	if err != nil {
		panic(err.Error)
	}

	serialised, err := json.Marshal(data)
	if err != nil {
		f.Close()
		panic(err.Error)
	}
	_, err = f.Write(serialised)
	if err != nil {
		f.Close()
		panic(err.Error)
	}
	f.Close()
}

func (c *Calibrate) importCalibrationMessage(msg CalibrationMessage, lit []int32, patternNumber int, capture int, wg *sync.WaitGroup) {
	r := c.convertCalibrationMessage(msg, lit)
	c.rawData = append(c.rawData, r)
	c.storeRawData(r, patternNumber, capture)
	wg.Done()
}

func (c *Calibrate) CalculateFrame(runtimeMs int64) (*Frame) {
	return c.frame
}

func (c *Calibrate) Subscribe() {
	if token := c.client.Subscribe(c.config.Mqtt.Topics.CalibrateClient, 0, c.handleClientMessages); token.Wait() && token.Error() != nil {
        log.Println(token.Error())
        os.Exit(1)
    }
}
