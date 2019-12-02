package stream

import (
	"math/rand"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/lucasb-eyer/go-colorful"
)

// A Twinkle is an Animation that twinkles random particles.
type Twinkle struct {
	client mqtt.Client
	numParticles int
	backColour colorful.Color

	initialised bool
	particles map[int]bool
}

// NewTwinkle creates an instance of a Twinkle object.
func NewTwinkle(client mqtt.Client, numParticles int, backColour colorful.Color) (*Twinkle) {
	t := new(Twinkle)
	t.client = client
	t.numParticles = numParticles
	t.backColour = backColour

	t.initialised = false
	t.particles = make(map[int]bool)

	return t
}

// CalculateFrame creates a new Frame instance.
func (t *Twinkle) CalculateFrame() (*Frame) {
	f := NewFrame()
	numPixels := len(f.pixels)
	if !t.initialised {
		for i := 0; i < t.numParticles; i++ {
			t.particles[rand.Intn(numPixels)] = true
		}
		t.initialised = true
	}

	for i := 0; i < numPixels; i++ {
		_, found := t.particles[i]
		var c colorful.Color
		if found {
			c, _ = colorful.Hex("#404040")
		} else {
			c = t.backColour
		}
		f.pixels[i] = c
	}

	return f
}
