package stream

import (
	"math/rand"

	"github.com/lucasb-eyer/go-colorful"
)

type multiParticle struct {
	lut        []float64
	lutLength  int
	current    int
	running    bool
	colour     colorful.Color
	NextColour colorful.Color
}

func newMultiParticle(colour colorful.Color, lut []float64) *multiParticle {
	p := new(multiParticle)

	p.colour = colour
	p.NextColour = colour
	p.lut = lut
	p.lutLength = len(lut)
	p.current = 0
	p.running = false

	return p
}

func (p *multiParticle) increment() {
	if p.running {
		p.current++
		if p.current > len(p.lut)/2 {
			p.colour = p.NextColour
		}

		if p.current == len(p.lut)-1 {
			p.current = 0
			p.running = false
		}
	}
}

func (p *multiParticle) scintillate() bool {
	result := !p.running
	p.running = true
	return result
}

func (p *multiParticle) currentColour() colorful.Color {
	if p.running {
		gain := p.lut[p.current]
		h, c, l := p.colour.Hcl()

		// Calculate the difference to the max luminance we want
		lumDiff := 0.6 - l

		return colorful.Hcl(h, c, l+(lumDiff*gain))
	} else {
		return p.colour
	}
}

// A MultiTwinkle is an Animation that twinkles random particles.
type MultiTwinkle struct {
	lut                 []float64
	backColours         []colorful.Color
	runtimeMs           int64
	scintillationChance int32
	pixels              []*multiParticle
}

// NewMultiTwinkle creates an instance of a Twinkle object.
func NewMultiTwinkle(scintillationChance int32, backColours []colorful.Color, runtimeMs int64) *MultiTwinkle {

	t := new(MultiTwinkle)
	t.lut = make([]float64, 0, 13)
	t.lut = append(t.lut, 0.05, 0.1, 0.2, 0.3, 0.6, 0.9, 1.0, 0.9, 0.6, 0.3, 0.2, 0.1, 0.05)
	t.backColours = backColours
	t.scintillationChance = scintillationChance
	t.pixels = nil

	return t
}

func (t *MultiTwinkle) getRandomBackColour() colorful.Color {
	return t.backColours[rand.Int31n(int32(len(t.backColours)))]
}

// CalculateFrame creates a new Frame instance.
func (t *MultiTwinkle) CalculateFrame(runtimeMs int64) *Frame {
	t.runtimeMs = runtimeMs

	f := NewFrame()
	numPixels := len(f.pixels)

	// Initialise if we need to
	if t.pixels == nil {
		t.pixels = make([]*multiParticle, numPixels)
		for i := 0; i < numPixels; i++ {
			t.pixels[i] = newMultiParticle(t.getRandomBackColour(), t.lut)
		}
	}

	for i := 0; i < numPixels; i++ {
		// Start scintillation by chance
		if rand.Int31n(t.scintillationChance) == 0 {
			if t.pixels[i].scintillate() {
				t.pixels[i].NextColour = t.getRandomBackColour()
			}
		}

		// Always increment, it'll only affect those pixels that are scintillating
		t.pixels[i].increment()

		f.pixels[i] = t.pixels[i].currentColour()
	}

	return f
}