package stream

import (
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Streamer that streams RGB data frames to an ledrx device.
type Streamer struct {
	config      Config
	client      mqtt.Client
	calibrate   *Calibrate
	animation   Animation
	frameTimeMs int64
	runtimeMs   int64
}

// NewStreamer creates an instance of a Streamer.
func NewStreamer(config Config, client mqtt.Client) *Streamer {
	s := new(Streamer)
	s.config = config
	s.client = client
	s.frameTimeMs = 21
	s.runtimeMs = 0

	// Use a controller as the animation, internally it will control multiple animations
	s.calibrate = NewCalibrate(s.config, s.client)
	frameRate := 1000.0 / float64(s.frameTimeMs)
	log.Printf("Frame rate: %0.1f fps", frameRate)
	c := NewController(s.runtimeMs, frameRate, 30*time.Second, s.calibrate)
	s.animation = c
	go c.Run() // The controller has a timer that needs to be started

	return s
}

// SendFrame sends a frame as binary over MQTT to an ledrx device.
func (s *Streamer) SendFrame() {
	s.runtimeMs += s.frameTimeMs
	f := s.animation.CalculateFrame(s.runtimeMs)

	// The animation can opt to not send a frame by returning nil
	if f != nil {
		b, _ := f.MarshalBinary()
		token := s.client.Publish(s.config.Mqtt.Topics.Stream, 0, false, b)
		token.Wait()
	}
}

// Run causes the Streamer to send Frames continuously.
func (s *Streamer) Run() {
	publishTimer := time.NewTicker(time.Duration(s.frameTimeMs) * time.Millisecond)
	for {
		<-publishTimer.C
		s.SendFrame()
	}
}

func (s *Streamer) Subscribe() {
	// Register for calibration requests
	s.calibrate.Subscribe()
}
