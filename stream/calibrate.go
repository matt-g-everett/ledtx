package stream


import (
	"encoding/json"
	"log"
	"math"
	"os"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/lucasb-eyer/go-colorful"
)

type CalibrationMessage struct {
	Type string `json:"type"`
	Locations []float64 `json:"locations"`
}

type Point struct {
	x float64
	y float64
}

type Calibrate struct {
	config Config
	client mqtt.Client
	C chan bool
	abort chan bool
	started bool

	litLength int
	lit map[int]bool
	iteration int
	bins map[Point]map[int]int
	resolved map[int]bool
	frame *Frame
	msg chan CalibrationMessage
	remainder int
}

func NewCalibrate(config Config, client mqtt.Client) *Calibrate {
	c := new(Calibrate)
	c.config = config
	c.client = client
	c.C = make(chan bool)
	c.started = false
	return c
}

func (c *Calibrate) prepareFrame() {
	c.lit = make(map[int]bool)
	for i := 0; i < len(c.frame.pixels); i++ {
		litlen := 1 << c.litLength
		lit := (int(math.Floor(float64(i) / float64(litlen))) % 2) == c.remainder
		c.lit[i] = lit
		if lit {
			c.frame.pixels[i], _ = colorful.Hex("#404040")
		} else {
			c.frame.pixels[i], _ = colorful.Hex("#000000")
		}
	}
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
	c.abort = make(chan bool, 1)
	c.frame = NewFrame()
	pixelCount := len(c.frame.pixels)
	c.litLength = int(math.Ceil(math.Log2(float64(pixelCount)))) - 1
	log.Printf("lit length: %d", c.litLength)
	c.bins = make(map[Point]map[int]int)
	c.resolved = make(map[int]bool)
	for i := 0; i < pixelCount; i++ {
		c.resolved[i] = false
	}
	c.msg = make(chan CalibrationMessage)
	c.prepareFrame()
	c.C<- true

	time.Sleep(2000 * time.Millisecond)
	c.remainder = 0
	for ; c.litLength >= 0 && c.started; {
		c.prepareFrame()
		time.Sleep(200 * time.Millisecond)

		for i := 0; i < 10; i++ {
			token := c.client.Publish(c.config.Mqtt.Topics.CalibrateServer, 0, false, "snapshot")
			token.Wait()
			msg := <-c.msg

			c.storeLocations(msg)
		}

		if c.remainder == 1 {
			c.litLength--
		}
		c.remainder = c.remainder ^ 1
		log.Printf("REMAINDER: %d", c.remainder)
	}

	log.Printf("Count: %d", len(c.bins))
	//count := 0
	report := make(map[int]Point)
	reportCount := make(map[int]int)
	// for k, v := range c.bins {
	// 	highestCount := 0
	// 	highestPixel := -1
	// 	for pixel, count := range v {
	// 		if count > highestCount {
	// 			highestPixel = pixel
	// 			highestCount = count
	// 		}
	// 	}

	// 	if highestCount > 20 {
	// 		report[highestPixel] = k
	// 		count++
	// 	}
	// }


	for i := 0; i < pixelCount; i++ {
		highestCount := 0
		highestBin := Point{0, 0}
		for k, v := range c.bins {
			if v[i] > highestCount {
				highestBin = k
				highestCount = v[i]
			}
		}

		report[i] = highestBin
		reportCount[i] = highestCount
	}

	for i := 0; i < pixelCount; i++ {
		loc, found := report[i]
		count, _ := reportCount[i]
		if found {
			log.Printf("%d: %v (%d)", i, loc, count)
		} else {
			log.Printf("%d:", i)
		}
	}
}

func isBin(x float64, y float64, bin Point, threshold float64) bool {
	return math.Sqrt(math.Pow(math.Abs(x - bin.x), 2.0) + math.Pow(math.Abs(y - bin.y), 2.0)) < threshold
}

func (c *Calibrate) storeLocations(msg CalibrationMessage) {
	toAdd := make(map[Point]map[int]int)
	for i := 0; i < len(msg.Locations); i += 2 {
		found := false
		for k, v := range c.bins {
			if isBin(msg.Locations[i], msg.Locations[i+1], k, 3.0) {
				found = true
				for j := 0; j < len(c.lit); j++ {
					if c.lit[j] {
						v[j]++
					} else {
						v[j]--
					}
				}
			}
		}

		if !found {
			// Add the bin
			pixelMap := make(map[int]int)
			toAdd[Point{msg.Locations[i], msg.Locations[i+1]}] = pixelMap
			for j := 0; j < len(c.lit); j++ {
				if c.lit[j] {
					pixelMap[j] = 1
				} else {
					pixelMap[j] = -1
				}
			}
		}
	}

	for k, v := range toAdd {
		c.bins[k] = v
	}
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
