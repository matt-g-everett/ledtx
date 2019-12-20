package stream

import (
	"log"
	"math/rand"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

// Controller that manages animations.
type Controller struct {
	calibrate *Calibrate
	animationIndex int
	animationPlaylist []string
	animation Animation
	nextAnimation Animation
	animationTime time.Duration
	cycling bool
	runtimeMs int64
	frameRate float64
	transition float64
	transitionTimeSecs float64
	transitionIncrement float64
	rainbowGradient GradientTable
}

// NewController creates an instance of a Controller.
func NewController(runtimeMs int64, frameRate float64, animationTime time.Duration,
	calibrate *Calibrate) *Controller {

	c := new(Controller)

	c.rainbowGradient = GradientTable{
		{0.0, 0.0},
		{6.0, 0.04}, // Pink
		{87.0, 0.14}, // Red
		{88.0, 0.28}, // Orange
		{98.0, 0.42}, // Yellow
		{180.0, 0.56}, // Green
		{190.0, 0.70}, // Turquiose
		{320.0, 0.84}, // Blue
		{328.0, 0.91}, // Violet
		{360.0, 1.0}, // Pink wrap
	}

	c.animation = nil
	c.nextAnimation = nil
	c.calibrate = calibrate
	c.animationTime = animationTime
	c.cycling = true

	c.runtimeMs = runtimeMs
	c.frameRate = frameRate
	c.transition = 0.0
	c.transitionTimeSecs = 5.0
	c.transitionIncrement = 1.0 / (c.frameRate * c.transitionTimeSecs)

	c.animationPlaylist = []string{
		"twinkle:random",
		"twinkle:blue",
		"rainbow:random",
		"twinkle:random",
		"twinkle:pink",
		"rainbow:normal",
		"twinkle:gold",
		"twinkle:random",
		"twinkle:silver",
		"rainbow:random",
	}
	c.animationIndex = 0
	c.animation = c.getAnimation()

	return c
}

func (c *Controller) CalculateFrame(runtimeMs int64) (*Frame) {
	var f *Frame
	c.runtimeMs = runtimeMs
	if c.nextAnimation != nil {
		f1 := c.animation.CalculateFrame(runtimeMs)
		f2 := c.nextAnimation.CalculateFrame(runtimeMs)
		f = f1.InterpolateFrame(f2, c.transition)
		c.transition += c.transitionIncrement

		if c.transition >= 1.0 {
			c.animation = c.nextAnimation
			c.nextAnimation = nil
			c.transition = 0.0
		}
	} else {
		f = c.animation.CalculateFrame(runtimeMs)
	}

	return f
}

func (c *Controller) createKnownTwinkle(foreColour colorful.Color, backColour colorful.Color) Animation {
	return NewTwinkle(rand.Int31n(600) + 200, foreColour, backColour, c.runtimeMs)
}

func (c *Controller) createRandomTwinkle(foreColour colorful.Color) Animation {
	return NewTwinkle(
		rand.Int31n(900) + 100,
		foreColour,
		colorful.Hsl(rand.Float64() * 360.0, 1.0, 0.02), c.runtimeMs)
}

func (c *Controller) createKnownRainbow() Animation {
	return NewGradientTrail(c.rainbowGradient, 180, 0.06, c.runtimeMs, -0.03)
}

func (c *Controller) createRandomRainbow() Animation {
	speed := (rand.Float64() - 0.5) / 6.0
	trailLength := rand.Int31n(970) + 30
	return NewGradientTrail(c.rainbowGradient, uint32(trailLength), 0.06, c.runtimeMs, speed)
}

func (c *Controller) getAnimation() Animation {
	twinkleHighlight, _ := colorful.Hex("#808080")

	var animation Animation
	switch c.animationPlaylist[c.animationIndex] {
	case "twinkle:blue":
		blue, _ := colorful.Hex("#000005")
		animation = c.createKnownTwinkle(twinkleHighlight, blue)
	case "twinkle:pink":
		pink, _ := colorful.Hex("#100505")
		animation = c.createKnownTwinkle(twinkleHighlight, pink)
	case "twinkle:random":
		animation = c.createRandomTwinkle(twinkleHighlight)
	case "twinkle:silver":
		silver, _ := colorful.Hex("#030303")
		animation = c.createKnownTwinkle(twinkleHighlight, silver)
	case "twinkle:gold":
		silver, _ := colorful.Hex("#050401")
		animation = c.createKnownTwinkle(twinkleHighlight, silver)
	case "rainbow:normal":
		animation = c.createKnownRainbow()
	case "rainbow:random":
		animation = c.createRandomRainbow()
	}

	return animation
}

func (c *Controller) cycleAnimation() {
	if c.cycling {
		c.animationIndex++
		c.animationIndex %= len(c.animationPlaylist)
		log.Printf("Cycling to %s", c.animationPlaylist[c.animationIndex])
		c.nextAnimation = c.getAnimation()
	}
}

// Run causes the Controller to cycle through animations.
func (c *Controller) Run() {
	publishTimer := time.NewTicker(c.animationTime)
	for {
		select {
		case <-publishTimer.C:
			c.cycleAnimation()
		case start, _ := <-c.calibrate.C:
			if start {
				c.cycling = false
				c.animation = c.calibrate
			} else {
				c.cycling = true
				c.cycleAnimation()
			}
		}
	}
}
