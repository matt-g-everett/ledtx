package stream

import (
	"container/list"
	"math"
	"math/rand"

	"github.com/fogleman/ease"
	"github.com/lucasb-eyer/go-colorful"
)

type streakParticle struct {
	colour    colorful.Color
	start     float64
	current   float64
	increment float64
	length    float64
	gainRate  float64
	easingOut bool
}

func newStreakParticle() *streakParticle {
	p := new(streakParticle)
	p.colour = colorful.Color{R: 0.45, G: -0.54, B: 0.02}
	p.start = 0
	p.current = 0
	p.increment = 0.2
	p.length = 10
	p.gainRate = 0.05
	p.easingOut = false
	return p
}

func (p *streakParticle) incrementPosition(numPixels float64) bool {
	p.current += p.increment
	if p.current > numPixels {
		return false
	} else if p.current < 0-p.length {
		return false
	}

	return true
}

func (p *streakParticle) calcEaseDistance() float64 {
	return math.Abs(p.current-p.start) * p.gainRate
}

func (p *streakParticle) isLive(easeDistance float64) bool {
	return easeDistance <= 2
}

func (p *streakParticle) overallGain(easeDistance float64) float64 {
	if easeDistance > 2 {
		return 0
	} else if easeDistance > 1 {
		easeDistance = 1 - (easeDistance - 1)
	}

	return ease.InOutQuad(easeDistance)
}

func (p *streakParticle) addStreak(frame *Frame) bool {
	easeDistance := p.calcEaseDistance()
	live := p.isLive(easeDistance)
	bias := p.overallGain(easeDistance)
	if live {
		start := int(math.Ceil(p.current))
		end := int(math.Floor(p.current + p.length))
		for i := start; i <= end; i++ {
			frame.pixels[i].BlendHcl(p.colour, bias)
		}
	}

	return live
}

// A Streak is an Animation that creates streaks across the tree that fade in then out.
type Streak struct {
	backColour   colorful.Color
	runtimeMs    int64
	streakChance int32
	particles    *list.List
}

// NewStreak creates an instance of a Streak object.
func NewStreak(runtimeMs int64, streakChance int32, backColour colorful.Color) *Streak {
	t := new(Streak)
	t.streakChance = streakChance
	t.backColour = backColour
	t.runtimeMs = runtimeMs
	t.particles = list.New()

	return t
}

// CalculateFrame creates a new Frame instance.
func (s *Streak) CalculateFrame(runtimeMs int64) *Frame {
	s.runtimeMs = runtimeMs

	f := NewFrame()
	numPixels := len(f.pixels)
	for i := 0; i < numPixels; i++ {
		f.pixels[i] = s.backColour
	}

	toDelete := make([]*list.Element, 0, s.particles.Len())
	for e := s.particles.Front(); e != nil; e = e.Next() {
		particle, _ := e.Value.(*streakParticle)
		more := particle.incrementPosition(float64(numPixels))
		if more {
			more = particle.addStreak(f)
		}

		if !more {
			toDelete = append(toDelete, e)
		}
	}

	if rand.Int31n(s.streakChance) == 0 {
		// Create a randomised new particle
		p := newStreakParticle()
		s.particles.PushBack(p)
	}

	for _, e := range toDelete {
		s.particles.Remove(e)
	}

	return f
}
