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
	Locations []float32 `json:"locations"`
}

type Point struct {
	x float32
	y float32
}

type Calibrate struct {
	config Config
	client mqtt.Client
	CalStart chan bool
	started bool

	litLength int
	iteration int
	bins map[Point][]int
	resolved map[int]bool
	frame *Frame
	msg chan CalibrationMessage
}

func NewCalibrate(config Config, client mqtt.Client) *Calibrate {
	c := new(Calibrate)
	c.config = config
	c.client = client
	c.CalStart = make(chan bool)
	c.started = false
	return c
}

func (c *Calibrate) prepareFrame() {
	for i := 0; i < len(c.frame.pixels); i++ {
		litlen := 1<<c.litLength
		lit := (int(math.Floor(float64(i) / float64(litlen))) % 2) < 1
		if lit {
			c.frame.pixels[i], _ = colorful.Hex("#404040")
		} else {
			c.frame.pixels[i], _ = colorful.Hex("#000000")
		}
	}
}

func (c *Calibrate) handleClientMessages(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received msg %d on %s: %s\n", msg.MessageID(), msg.Topic(), msg.Payload())

	var message CalibrationMessage
	json.Unmarshal(msg.Payload(), &message)


	if (message.Type == "start" && c.started == false) {

		go c.runCalibration()
	} else if c.started {
		c.msg<- message
	}

	//token := c.client.Publish(c.config.Mqtt.Topics.CalibrateServer, 0, false, "go")
	//token.Wait()
}

func (c *Calibrate) runCalibration() {
	c.frame = NewFrame()
	pixelCount := len(c.frame.pixels)
	c.litLength = int(math.Ceil(math.Log2(float64(pixelCount))))
	log.Printf("lit length: %d", c.litLength)
	c.bins = make(map[Point][]int)
	c.resolved = make(map[int]bool)
	for i := 0; i < pixelCount; i++ {
		c.resolved[i] = false
	}
	c.msg = make(chan CalibrationMessage)
	c.prepareFrame()
	c.CalStart<- true

	for ; c.litLength >= 0; c.litLength-- {
		c.prepareFrame()
		time.Sleep(200 * time.Millisecond)

		// select {
		// case msg, _ := <-c.msg:
		// 	log.Printf("%+v", msg)
		// }

		// if pixelCount == 1 {
		// 	break
		// }
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
