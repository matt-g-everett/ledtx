package stream


import (
	"log"
	"os"

	"github.com/eclipse/paho.mqtt.golang"
)

type Calibrate struct {
	config Config
	client mqtt.Client
	C chan bool
}

func NewCalibrate(config Config, client mqtt.Client) *Calibrate {
	c := new(Calibrate)
	c.config = config
	c.client = client
	return c
}

func (c *Calibrate) handleClientMessages(client mqtt.Client, msg mqtt.Message) {
    log.Printf("Received msg %d on %s: %s\n", msg.MessageID(), msg.Topic(), msg.Payload())
}

func (c *Calibrate) Subscribe() {
	if token := c.client.Subscribe(c.config.Mqtt.Topics.CalibrateClient, 0, c.handleClientMessages); token.Wait() && token.Error() != nil {
        log.Println(token.Error())
        os.Exit(1)
    }
}
