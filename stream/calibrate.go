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

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	binSimilarityDistance float64 = 4.0
	binHitThreshold       int32   = 1
)

// Calibrate finds the locations of LEDs in polar co-ordinates where the radius is the distance projected
// onto the cone-shaped tree from top to bottom
type Calibrate struct {
	config         Config
	client         mqtt.Client
	C              chan bool
	started        bool
	iteration      int
	bins           map[Point]map[int]int
	onscreenFrame  *Frame
	offscreenFrame *Frame
	ackID          uint8
	ackChan        chan AckMessage
	dataChan       chan DataMessage
	rawData        []*RawCalibrationData
	aggregated     *AggregatedData

	binWriteLock sync.Mutex
}

// NewCalibrate creates an instance of a Calibrate struct
func NewCalibrate(config Config, client mqtt.Client) *Calibrate {
	c := new(Calibrate)
	c.config = config
	c.client = client
	c.C = make(chan bool)
	c.started = false
	c.ackID = 1

	c.onscreenFrame = NewFrame()
	c.offscreenFrame = NewFrame()

	// Turn all the lights on to start calibration
	c.showCalibrationFrame(1, 0)

	return c
}

func (c *Calibrate) incrementAckID() {
	c.ackID++

	// Avoid zero when wrapping because that means "don't ack"
	if c.ackID == 0 {
		c.ackID = 1
	}
}

func (c *Calibrate) prepareFS() {
	os.RemoveAll("caldata")
	os.MkdirAll("caldata/raw", 0755)
	os.MkdirAll("caldata/pixels", 0755)
}

func (c *Calibrate) showCalibrationFrame(interval int, offset int) []int32 {
	pixelCount := len(c.offscreenFrame.pixels)
	lit := make([]int32, pixelCount, pixelCount)

	for i := 0; i < pixelCount; i++ {
		if (i-offset)%interval == 0 {
			c.offscreenFrame.pixels[i], _ = colorful.Hex("#202020")
			lit[i] = 1
		} else {
			c.offscreenFrame.pixels[i], _ = colorful.Hex("#000000")
			lit[i] = 0
		}
	}

	c.actionFrame(true)

	return lit
}

func (c *Calibrate) showStatusFrame(resolved []Pixel) {
	pixelCount := len(c.offscreenFrame.pixels)
	for i := 0; i < pixelCount; i++ {
		if resolved[i].Resolved {
			c.offscreenFrame.pixels[i], _ = colorful.Hex("#002000")
		} else {
			c.offscreenFrame.pixels[i], _ = colorful.Hex("#000000")
		}
	}

	c.actionFrame(true)
}

func (c *Calibrate) actionFrame(switchFrames bool) {
	// Use the current ackID
	c.offscreenFrame.ackID = c.ackID

	if switchFrames {
		tempFrame := c.onscreenFrame
		c.onscreenFrame = c.offscreenFrame
		c.offscreenFrame = tempFrame
	}

	// The next frame will have a new ACK ID
	c.incrementAckID()
}

func (c *Calibrate) handleClientMessages(client mqtt.Client, msg mqtt.Message) {
	var message CalibrationMessage
	json.Unmarshal(msg.Payload(), &message)

	if !c.started && message.Type == "start" {
		go c.runCalibration()
	} else if message.Type == "ack" {
		var ackMsg AckMessage
		json.Unmarshal(msg.Payload(), &ackMsg)
		fmt.Printf("Got ACK message %s\n", string(msg.Payload()))
		c.ackChan <- ackMsg
	} else if c.started && message.Type == "data" {
		var dataMsg DataMessage
		fmt.Printf("Got data message %s\n", string(msg.Payload()))
		json.Unmarshal(msg.Payload(), &dataMsg)
		c.dataChan <- dataMsg
	}
}

func (c *Calibrate) resolve(aggregated *AggregatedData, resolved []Pixel) {
	for _, bin := range aggregated.Bins {
		var maxPixelFrequency int32 = 0 // Start with zero, ignore negative counts
		pixel := -1
		unique := true
		for i, p := range bin.Pixels {
			if p > maxPixelFrequency {
				maxPixelFrequency = p
				pixel = i
				unique = true
			} else if p == maxPixelFrequency {
				unique = false
			}
		}

		if pixel > -1 && maxPixelFrequency > 0 && unique {
			if !resolved[pixel].Resolved {
				resolved[pixel].Location = bin.Location
				resolved[pixel].Resolved = true
			}
		}
	}
}

func (c *Calibrate) runCalibration() {
	c.started = true
	pixelCount := len(c.offscreenFrame.pixels)
	c.aggregated = &AggregatedData{Bins: make([]*Bin, 0, 5000)}
	c.ackChan = make(chan AckMessage)
	c.dataChan = make(chan DataMessage)
	intervals := []int{1, 2, 3, 5, 7, 11, 13, 17, 19}
	rawDataCount := 0
	for _, interval := range intervals {
		rawDataCount += interval
	}

	c.rawData = make([]*RawCalibrationData, 0, rawDataCount)

	// Tell the controller that we're ready to start showing frames
	c.C <- true

	// Allow the camera to adjust exposure
	time.Sleep(2 * time.Second)

	c.prepareFS()
	capture := 0
	var importWaitGroup sync.WaitGroup

	for _, interval := range intervals {
		for o := 0; o < interval; o++ {

			lit := c.showCalibrationFrame(interval, o)
			log.Printf("Sending initial ACK ID: %d", c.ackID)

			for iter := 0; iter < 5; iter++ {
				// Loop until we get a reply
				exitTimeout := time.NewTimer(30 * time.Second)
				for true {
					ackTimeout := time.NewTimer(time.Second)
					select {
					case msg := <-c.ackChan:
						log.Printf("Received ACK %d", msg.AckID)
						// TODO: Store the ACK ID and check it
						break
					case <-ackTimeout.C:
						log.Println("Timed-out waiting for ACK, incrementing the ackId")
						c.actionFrame(false)
						log.Printf("Sending retry ACK ID: %d", c.ackID)
					case <-exitTimeout.C:
						log.Println("Can't get an ACK from ledrx, giving up the calibration")
						// Tell the controller to stop the calibration
						c.C <- false
						return
					}
				}

				token := c.client.Publish(c.config.Mqtt.Topics.CalibrateServer, 0, false, "snapshot")
				if o == 0 {
					// The light level is different when the interval changes so wait for exposure adjustment
					time.Sleep(1 * time.Second)
				} else {
					time.Sleep(100 * time.Millisecond)
				}

				token.Wait()
				log.Println("Published snapshot")

				t := time.NewTimer(time.Second)
				select {
				case msg := <-c.dataChan:
					log.Println("Received snapshot")
					importWaitGroup.Add(1)
					go c.importCalibrationMessage(msg, lit, capture, interval, o, &importWaitGroup)
				case <-t.C:
					log.Println("Message timed-out, retrying")
				}

				capture++
			}
		}
	}

	importWaitGroup.Wait()
	log.Println("########## DONE CAPTURING")

	c.aggregate()
	log.Printf("Bin count (total): %d", len(c.aggregated.Bins))

	highHitData := &AggregatedData{Bins: make([]*Bin, 0, 10000)}
	for _, b := range c.aggregated.Bins {
		if b.Hits >= binHitThreshold {
			highHitData.Bins = append(highHitData.Bins, b)
		}
	}
	log.Printf("Bin count (hits): %d", len(highHitData.Bins))

	c.store(c.aggregated, "caldata/aggregated_raw.json")
	c.store(highHitData, "caldata/aggregated.json")

	resolved := make([]Pixel, pixelCount, pixelCount)
	c.resolve(c.aggregated, resolved)
	c.store(resolved, "caldata/resolved.json")
	log.Println("Resolved")

	c.showStatusFrame(resolved)

	token := c.client.Publish(c.config.Mqtt.Topics.CalibrateServer, 0, false, "snapshot")
	token.Wait()
	log.Println("Published resolved")
}

func (c *Calibrate) aggregate() {
	// var wg sync.WaitGroup
	for _, r := range c.rawData {
		for _, l := range r.Locations {
			//wg.Add(len(r.Locations))
			c.incrementBin(l, r.Pixels)
		}
	}

	//wg.Wait()
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
			c.aggregated.Bins = append(c.aggregated.Bins, &Bin{Location: binLocation, Pixels: litCopy, Hits: 1})
		}
		c.binWriteLock.Unlock()

		// If we found it then another goroutine added it, so just increment as usual
		if found {
			c.doIncrementBin(foundBin, lit)
		}
	}
}

func isBin(a Point, b Point, threshold float64) bool {
	return math.Sqrt(math.Pow(math.Abs(a.X-b.X), 2.0)+math.Pow(math.Abs(a.Y-b.Y), 2.0)) < threshold
}

func (c *Calibrate) convertCalibrationMessage(msg DataMessage, lit []int32) *RawCalibrationData {
	pointCount := len(msg.Locations) / 2
	r := &RawCalibrationData{
		Pixels:    lit,
		Locations: make([]Point, pointCount, pointCount),
	}

	for i := 0; i < pointCount; i++ {
		r.Locations[i] = Point{
			X: msg.Locations[i*2],
			Y: msg.Locations[(i*2)+1],
		}
	}

	return r
}

func (c *Calibrate) store(data interface{}, filePath string) {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		panic(err.Error)
	}

	serialised, err := json.MarshalIndent(data, "", "  ")
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

func (c *Calibrate) importCalibrationMessage(msg DataMessage, lit []int32, capture int, interval int,
	offset int, wg *sync.WaitGroup) {

	r := c.convertCalibrationMessage(msg, lit)
	c.rawData = append(c.rawData, r)
	c.store(r, fmt.Sprintf("caldata/raw/raw-%03d-%02d-%02d.json", capture, interval, offset))
	wg.Done()
}

// CalculateFrame gets the onscreen frame
func (c *Calibrate) CalculateFrame(runtimeMs int64) *Frame {
	return c.onscreenFrame
}

// Subscribe to listen to the mobile app
func (c *Calibrate) Subscribe() {
	if token := c.client.Subscribe(c.config.Mqtt.Topics.CalibrateClient, 0, c.handleClientMessages); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
		os.Exit(1)
	}
}
