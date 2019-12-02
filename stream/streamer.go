package stream

import (
	"math"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

// Streamer that streams RGB data frames to an ledrx device.
type Streamer struct {
	client mqtt.Client
	rainbow GradientTable
	current float64
	trailLength int
}

// NewStreamer creates an instance of a Streamer.
func NewStreamer(client mqtt.Client) *Streamer {
	s := new(Streamer)
	s.client = client
	s.current = 0
	s.trailLength = 250
	s.rainbow = GradientTable{
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

	return s
}

// SendFrame sends a frame as binary over MQTT to an ledrx device.
func (s *Streamer) SendFrame() {
	f := NewFrame()
	saturation := 1.0
	luminance := 0.05
	numPixels := len(f.pixels)
	for i := 0; i < numPixels; i++ {
		t := math.Mod((float64(i + numPixels) - s.current), float64(s.trailLength)) / float64(s.trailLength)
		c := s.rainbow.GetColor(t, saturation, luminance)
		f.pixels[i] = c
	}

	s.current += 2.0
	s.current = math.Mod(s.current, float64(s.trailLength))

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
