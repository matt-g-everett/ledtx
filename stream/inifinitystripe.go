package stream

import "github.com/matt-g-everett/ledtx/stream/stripe"

// A GradientTrail is an Animation that cycles a gradient along an led strip.
type InfinityStripe struct {
	stripes     []stripe.Stripe
	current     float64
	runtimeMs   int64
	pixelsPerMs float64
	adjusted    bool
	generator   stripe.StripeGenerator
}

// NewInfinityStripe creates an instance of a InfinityStripe object.
func NewInfinityStripe(startTimeMs int64, pixelsPerMs float64, stripeGenerator stripe.StripeGenerator) *InfinityStripe {

	s := new(InfinityStripe)
	s.stripes = make([]stripe.Stripe, 0, 20)
	s.runtimeMs = startTimeMs
	s.pixelsPerMs = pixelsPerMs
	s.adjusted = true
	s.current = 0
	s.generator = stripeGenerator

	return s
}

func (s *InfinityStripe) addStripe() stripe.Stripe {
	stripe := s.generator.CreateStripe()
	s.stripes = append(s.stripes, stripe)
	return stripe
}

func (s *InfinityStripe) getStripe(offset float64) (stripe.Stripe, float64) {
	if len(s.stripes) == 0 {
		s.addStripe()
	}

	if offset < float64(s.stripes[0].Length) {
		firstStripe := s.stripes[0]
		return firstStripe, float64(firstStripe.Length)
	}

	var length int32 = 0
	for _, stripe := range s.stripes {
		length += stripe.Length
		if offset < float64(length) {
			return stripe, float64(length)
		}
	}

	for offset > float64(length) {
		stripe := s.addStripe()
		length += stripe.Length
	}

	lastStripe := s.stripes[len(s.stripes)-1]
	return lastStripe, float64(length)
}

// CalculateFrame creates a new Frame instance.
func (s *InfinityStripe) CalculateFrame(runtimeMs int64) *Frame {
	f := NewFrame()
	numPixels := len(f.pixels)

	// Cull stripes that have passed
	// toRemove := 0
	// for _, stripe := range s.stripes {
	// 	if s.current > float64(stripe.length) {
	// 		toRemove++
	// 		s.current -= float64(stripe.length)
	// 	}
	// }

	// if toRemove > 0 {
	// 	s.stripes = s.stripes[toRemove:]
	// }

	//maxOffset := 1.0 + 3.0*(float64(numPixels-1)/float64(numPixels))

	adjustmentFactor := 1.0
	currentStripe, stripeEnd := s.getStripe(s.current)
	for i := 0; i < numPixels; i++ {
		if s.adjusted {
			adjustmentFactor = 1.0 + 3.0*(float64(i)/float64(numPixels))
		}

		adjustedOffset := (adjustmentFactor * float64(i)) + s.current
		if adjustedOffset > stripeEnd {
			currentStripe, stripeEnd = s.getStripe(adjustedOffset)
		}

		f.pixels[i] = currentStripe.Colour
	}

	intervalMs := runtimeMs - s.runtimeMs
	s.runtimeMs = runtimeMs
	s.current += s.pixelsPerMs * float64(intervalMs)

	return f
}
