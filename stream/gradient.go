package stream

import (
	"github.com/lucasb-eyer/go-colorful"
)

// GradientTable stores a look-up table of colours interpolated by hue.
type GradientTable []struct {
	Hue float64
	Pos float64
}

// GetColor gets a colour at the specified point on the look-up table.
func (g GradientTable) GetColor(t, s, l float64) colorful.Color {
	for i := 0; i < len(g)-1; i++ {
		c1 := g[i]
		c2 := g[i+1]
		if c1.Pos <= t && t <= c2.Pos {
			// We are in between c1 and c2. Go blend them!
			h := (((t - c1.Pos) / (c2.Pos - c1.Pos)) * (c2.Hue - c1.Hue)) + c1.Hue
			return colorful.Hcl(h, s, l)
		}
	}

	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return colorful.Hcl(g[len(g)-1].Hue, 1.0, 0.05)
}
