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
}

// NewGradientTrail creates an instance of a GradientTrail object.
func NewGradientTrail(client mqtt.Client, gradient GradientTable, trailLength int) (*GradientTrail) {
	g := new(GradientTrail)
	g.client = client
	g.gradient = gradient
	g.trailLength = trailLength
	g.current = 0

	return g
}

// CalculateFrame creates a new Frame instance.
func (g *GradientTrail) CalculateFrame() (*Frame) {
	f := NewFrame()
	saturation := 1.0
	luminance := 0.05
	numPixels := len(f.pixels)
	for i := 0; i < numPixels; i++ {
		t := math.Mod((float64(i + numPixels) - g.current), float64(g.trailLength)) / float64(g.trailLength)
		c := g.gradient.GetColor(t, saturation, luminance)
		f.pixels[i] = c
	}

	g.current += 2.0
	g.current = math.Mod(g.current, float64(g.trailLength))

	return f
}
