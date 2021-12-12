package stream

import (
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

// GradientTable stores a look-up table of colours interpolated by hue.
type GradientTable []struct {
	Hue        float64
	Saturation float64
	Pos        float64
}

// GetColor gets a colour at the specified point on the look-up table.
func (g GradientTable) GetColor(t float64, l float64) colorful.Color {
	for i := 0; i < len(g)-1; i++ {
		c1 := g[i]
		c2 := g[i+1]
		if c1.Pos <= t && t <= c2.Pos {
			// We are in between c1 and c2. Go blend them!

			hRatio := (t - c1.Pos) / (c2.Pos - c1.Pos)
			hRaw := (hRatio * (c2.Hue - c1.Hue)) + c1.Hue
			// Clamp hue to a positive number and wrap it
			h := math.Mod(math.Max(hRaw, 0.0), 360.0)

			sRatio := (t - c1.Pos) / (c2.Pos - c1.Pos)
			sRaw := sRatio*(c2.Saturation-c1.Saturation) + c1.Saturation
			// Clamp saturation between 0.0 and 1.0
			s := math.Max(math.Min(sRaw, 1.0), 0.0)
			return colorful.Hcl(h, s, l)
		}
	}

	// Nothing found? Means we're at (or past) the last gradient keypoint.
	last := g[len(g)-1]
	return colorful.Hcl(last.Hue, last.Saturation, l)
}

// GetColorFixedSaturation gets a colour at the specified point on the look-up table.
func (g GradientTable) GetColorFixedSaturation(t float64, s float64, l float64) colorful.Color {
	for i := 0; i < len(g)-1; i++ {
		c1 := g[i]
		c2 := g[i+1]
		if c1.Pos <= t && t <= c2.Pos {
			// We are in between c1 and c2. Go blend them!
			h := (((t - c1.Pos) / (c2.Pos - c1.Pos)) * (c2.Hue - c1.Hue)) + c1.Hue
			s := (((t - c1.Pos) / (c2.Pos - c1.Pos)) * (c2.Saturation - c1.Saturation)) + c1.Saturation
			return colorful.Hcl(h, s, l)
		}
	}

	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return colorful.Hcl(g[len(g)-1].Hue, s, l)
}
