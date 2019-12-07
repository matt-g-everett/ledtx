package stream

import (
	"math/rand"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/lucasb-eyer/go-colorful"
)

type scintillatingParticle struct {
	lut []float64
	current int
}

func newScintillatingParticle() *scintillatingParticle {
	p := new(scintillatingParticle)
	p.current = 0
	p.lut = make([]float64, 0, 9)
	p.lut = append(p.lut, 0.05, 0.1, 0.2, 0.3, 0.6, 0.9, 1.0, 0.9, 0.6, 0.3, 0.2, 0.1, 0.05)
	return p
}

func (p *scintillatingParticle) increment() bool {
	p.current++
	return p.current < len(p.lut)
}

func (p *scintillatingParticle) gain() float64 {
	return p.lut[p.current]
}


// A Twinkle is an Animation that twinkles random particles.
type Twinkle struct {
	client mqtt.Client
	numParticles int
	foreColour colorful.Color
	backColour colorful.Color
	runtimeMs int64
	scintillationChance int32

	particles map[int]*scintillatingParticle
}

// NewTwinkle creates an instance of a Twinkle object.
func NewTwinkle(client mqtt.Client, scintillationChance int32, foreColour colorful.Color,
	backColour colorful.Color, runtimeMs int64) (*Twinkle) {

	t := new(Twinkle)
	t.client = client
	t.foreColour = foreColour
	t.backColour = backColour
	t.scintillationChance = scintillationChance

	t.particles = make(map[int]*scintillatingParticle)

	return t
}

// CalculateFrame creates a new Frame instance.
func (t *Twinkle) CalculateFrame(runtimeMs int64) (*Frame) {
	t.runtimeMs = runtimeMs

	f := NewFrame()
	numPixels := len(f.pixels)
	for i := 0; i < numPixels; i++ {
		var color colorful.Color

		created := false
		p, found := t.particles[i]
		if !found && rand.Int31n(t.scintillationChance) == 0 {
			// Create the particle
			p = newScintillatingParticle()
			t.particles[i] = p
			created = true
		}

		if found || created {
			h, c, l := t.foreColour.Hcl()
			color = colorful.Hcl(h, c, l * p.gain())

			if !p.increment() {
				delete(t.particles, i)
			}
		} else {
			color = t.backColour
		}

		f.pixels[i] = color
	}

	return f
}
