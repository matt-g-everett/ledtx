package stream

import (
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/lucasb-eyer/go-colorful"
)

// Streamer that streams RGB data frames to an ledrx device.
type Streamer struct {
	client mqtt.Client
	animation Animation
}

// NewStreamer creates an instance of a Streamer.
func NewStreamer(client mqtt.Client) *Streamer {
	s := new(Streamer)
	s.client = client

	backColour, _ := colorful.Hex("#100505") // ("#000005")
	s.animation = NewTwinkle(s.client, 60, backColour)

	gradient := GradientTable{
		{0.0, 0.0},
		{6.0, 0.04}, // Pink
		{87.0, 0.14}, // Red
		{88.0, 0.28}, // Orange
		{98.0, 0.42}, // Yellow
		{180.0, 0.56}, // Green
		{190.0, 0.70}, // Turquiose
		{320.0, 0.84}, // Blue
		{328.0, 0.91}, // Violet
		{360.0, 1.0}, // Pink wrap
	}
	s.animation = NewGradientTrail(s.client, gradient, 200)



	return s
}

// SendFrame sends a frame as binary over MQTT to an ledrx device.
func (s *Streamer) SendFrame() {
	f := s.animation.CalculateFrame()
	b, _ := f.MarshalBinary()
	token := s.client.Publish("home/xmastree/stream", 2, false, b)
	token.Wait()
}

// Run causes the Streamer to send Frames continuously.
func (s *Streamer) Run() {
	publishTimer := time.NewTicker(33 * time.Millisecond)
	for {
		<-publishTimer.C
		s.SendFrame()
	}
}
