package stream

import (
	"math/rand"

	"github.com/lucasb-eyer/go-colorful"
)

type Stripe struct {
	colour colorful.Color
	length int32
}

// A GradientTrail is an Animation that cycles a gradient along an led strip.
type InfinityStripe struct {
	stripes       []Stripe
	current       float64
	runtimeMs     int64
	pixelsPerMs   float64
	adjusted      bool
	colours       []colorful.Color
	currentColour int
}

// NewInfinityStripe creates an instance of a InfinityStripe object.
func NewInfinityStripe(startTimeMs int64, pixelsPerMs float64) *InfinityStripe {

	s := new(InfinityStripe)
	s.stripes = make([]Stripe, 0, 20)
	s.runtimeMs = startTimeMs
	s.pixelsPerMs = pixelsPerMs
	s.adjusted = true
	s.current = 0

	s.colours = []colorful.Color{
		{R: 0.45, G: -0.54, B: 0.02},   // Pink
		{R: 0.23, G: 0.04, B: -0.87},   // Orange
		colorful.Hcl(280.0, 1.0, 0.06)} // Blue
	s.currentColour = 0

	return s
}

func (s *InfinityStripe) addStripe() Stripe {
	// colour := s.colours[s.currentColour]
	// s.currentColour++
	// s.currentColour = s.currentColour % len(s.colours)
	// stripe := Stripe{colour, 200}

	colour := colorful.Hsl(rand.Float64()*360.0, 1.0, 0.2)
	stripeLength := rand.Int31n(250) + 150
	stripe := Stripe{colour, stripeLength}

	s.stripes = append(s.stripes, stripe)
	return stripe
}

func (s *InfinityStripe) getStripe(offset float64) (Stripe, float64) {
	if len(s.stripes) == 0 {
		s.addStripe()
	}

	if offset < float64(s.stripes[0].length) {
		firstStripe := s.stripes[0]
		return firstStripe, float64(firstStripe.length)
	}

	var length int32 = 0
	for _, stripe := range s.stripes {
		length += stripe.length
		if offset < float64(length) {
			return stripe, float64(length)
		}
	}

	for offset > float64(length) {
		stripe := s.addStripe()
		length += stripe.length
	}

	lastStripe := s.stripes[len(s.stripes)-1]
	return lastStripe, float64(length)
}

// CalculateFrame creates a new Frame instance.
func (s *InfinityStripe) CalculateFrame(runtimeMs int64) *Frame {
	f := NewFrame()
	numPixels := len(f.pixels)

	// Cull stripes that have passed
	toRemove := 0
	for _, stripe := range s.stripes {
		if s.current > float64(stripe.length) {
			toRemove++
			s.current -= float64(stripe.length)
		}
	}

	if toRemove > 0 {
		s.stripes = s.stripes[toRemove:]
	}

	adjustmentFactor := 1.0
	currentStripe, stripeEnd := s.getStripe(s.current)
	for i := 0; i < numPixels; i++ {
		if s.adjusted {
			adjustmentFactor = 1.0 + 1.4*(float64(i)/float64(numPixels))
		}

		adjustedOffset := (adjustmentFactor * float64(i)) + s.current
		if adjustedOffset > stripeEnd {
			currentStripe, stripeEnd = s.getStripe(adjustedOffset)
		}

		f.pixels[i] = currentStripe.colour
	}

	intervalMs := runtimeMs - s.runtimeMs
	s.runtimeMs = runtimeMs
	s.current += s.pixelsPerMs * float64(intervalMs)

	return f
}
