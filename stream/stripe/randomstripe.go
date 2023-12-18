package stripe

import (
	"math/rand"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/matt-g-everett/ledtx/util"
)

type RandomStripeGenerator struct {
	palette       []colorful.Color
	current       int
	stripeMin     int32
	stripeMax     int32
	minSaturation float64
	maxSaturation float64
}

func NewRandomStripeGenerator(palette []colorful.Color) *RandomStripeGenerator {
	g := new(RandomStripeGenerator)
	g.palette = palette
	g.stripeMax = 1000
	g.stripeMin = 200
	return g
}

func NewRandomStripeGeneratorVariableSaturation(minSaturation float64, maxSaturation float64) *RandomStripeGenerator {
	g := new(RandomStripeGenerator)
	g.stripeMax = 1000
	g.stripeMin = 200
	g.minSaturation = minSaturation
	g.maxSaturation = maxSaturation
	return g
}

func (g *RandomStripeGenerator) CreateStripe() Stripe {
	var colour colorful.Color
	if g.palette == nil {
		colour = colorful.Hsl(rand.Float64()*360.0, util.RandomiseSaturation(g.minSaturation, g.maxSaturation), 0.2)
	} else {
		// Choose a new colour that's different from the previous colour
		for {
			newCurrent := rand.Intn(len(g.palette))
			if newCurrent != g.current {
				g.current = newCurrent
				break
			}
		}

		colour = g.palette[g.current]
	}

	stripeLength := rand.Int31n(g.stripeMax-g.stripeMin) + g.stripeMin
	return Stripe{colour, stripeLength}
}
