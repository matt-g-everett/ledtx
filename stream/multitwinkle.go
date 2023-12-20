package stream

import (
	"math/rand"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/matt-g-everett/ledtx/util"
)

type multiParticle struct {
	staticLut  []float64
	lut        []float64
	memoizer   util.Memoizer
	current    int
	running    bool
	colour     colorful.Color
	NextColour colorful.Color
}

func newMultiParticle(colour colorful.Color, lut []float64, memoizer util.Memoizer) *multiParticle {
	p := new(multiParticle)

	p.colour = colour
	p.NextColour = colour
	p.staticLut = lut
	p.lut = lut
	p.memoizer = memoizer
	p.current = 0
	p.running = false

	p.updateLut()

	return p
}

func (p *multiParticle) updateLut() {
	if p.staticLut == nil {
		p.lut = util.GenerateLutMemoized((rand.Intn(18)+6)*2, p.memoizer)
	}
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

			// Update the LUT every time we finish a scintillation
			p.updateLut()
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
	memoizer            util.Memoizer
}

// NewMultiTwinkle creates an instance of a Twinkle object.
func NewMultiTwinkle(scintillationChance int32, backColours []colorful.Color, lut []float64, runtimeMs int64) *MultiTwinkle {
	t := new(MultiTwinkle)

	t.lut = lut
	t.backColours = backColours
	t.scintillationChance = scintillationChance
	t.pixels = nil
	t.memoizer = util.Memoizer{}

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
			t.pixels[i] = newMultiParticle(t.getRandomBackColour(), t.lut, t.memoizer)
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
