package stream

import (
	"math"
)

// A GradientTrail is an Animation that cycles a gradient along an led strip.
type GradientTrail struct {
	gradient    GradientTable
	current     float64
	trailLength uint32
	luminance   float64
	runtimeMs   int64
	pixelsPerMs float64
	adjusted    bool
}

// NewGradientTrail creates an instance of a GradientTrail object.
func NewGradientTrail(gradient GradientTable, trailLength uint32,
	luminance float64, startTimeMs int64, pixelsPerMs float64) *GradientTrail {

	g := new(GradientTrail)
	g.gradient = gradient
	g.trailLength = trailLength
	g.luminance = luminance
	g.runtimeMs = startTimeMs
	g.pixelsPerMs = pixelsPerMs
	g.adjusted = true
	g.current = 0

	return g
}

// CalculateFrame creates a new Frame instance.
func (g *GradientTrail) CalculateFrame(runtimeMs int64) *Frame {
	adjustmentFactor := 1.0
	f := NewFrame()
	numPixels := len(f.pixels)
	for i := 0; i < numPixels; i++ {
		if g.adjusted {
			adjustmentFactor = 1.0 + 1.4*(float64(numPixels-i)/float64(numPixels))
		}

		adjustedTrailLength := float64(g.trailLength) * adjustmentFactor
		t := math.Mod(float64(i)+(adjustmentFactor*g.current), float64(adjustedTrailLength)) / float64(adjustedTrailLength)
		c := g.gradient.GetColor(t, g.luminance)
		f.pixels[i] = c
	}

	intervalMs := runtimeMs - g.runtimeMs
	g.runtimeMs = runtimeMs

	g.current += g.pixelsPerMs * float64(intervalMs)
	g.current = math.Mod(g.current, float64(g.trailLength))
	// Deal with wrapping of negative current
	if g.current < 0 {
		g.current = float64(g.trailLength) + g.current
	}

	return f
}
