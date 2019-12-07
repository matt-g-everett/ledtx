package stream

import (
	"math"

	"github.com/eclipse/paho.mqtt.golang"
)

// A GradientTrail is an Animation that cycles a gradient along an led strip.
type GradientTrail struct {
	client mqtt.Client
	gradient GradientTable
	current float64
	trailLength int
	luminance float64
	runtimeMs int64
	pixelsPerMs float64
}

// NewGradientTrail creates an instance of a GradientTrail object.
func NewGradientTrail(client mqtt.Client, gradient GradientTable, trailLength int,
	luminance float64, startTimeMs int64, pixelsPerMs float64) (*GradientTrail) {

	g := new(GradientTrail)
	g.client = client
	g.gradient = gradient
	g.trailLength = trailLength
	g.luminance = luminance
	g.runtimeMs = startTimeMs
	g.pixelsPerMs = pixelsPerMs
	g.current = 0

	return g
}

// CalculateFrame creates a new Frame instance.
func (g *GradientTrail) CalculateFrame(runtimeMs int64) (*Frame) {
	f := NewFrame()
	saturation := 1.0
	numPixels := len(f.pixels)
	for i := 0; i < numPixels; i++ {
		t := math.Mod((float64(i + numPixels) - g.current), float64(g.trailLength)) / float64(g.trailLength)
		c := g.gradient.GetColor(t, saturation, g.luminance)
		f.pixels[i] = c
	}

	intervalMs := runtimeMs - g.runtimeMs
	g.runtimeMs = runtimeMs
	g.current += g.pixelsPerMs * float64(intervalMs)
	g.current = math.Mod(g.current, float64(g.trailLength))

	return f
}
