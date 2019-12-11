package stream

import (
	"log"
	"reflect"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

// Controller that manages animations.
type Controller struct {
	calibrate *Calibrate
	animation Animation
	nextAnimation Animation
	animationTime time.Duration
	runtimeMs int64
	frameRate float64
	transition float64
	transitionTimeSecs float64
	transitionIncrement float64
}

// NewController creates an instance of a Controller.
func NewController(runtimeMs int64, frameRate float64, animationTime time.Duration,
	calibrate *Calibrate) *Controller {

	c := new(Controller)

	backColour, _ := colorful.Hex("#000005") //("#100505")
	foreColour, _ := colorful.Hex("#808080") //("#404040")
	c.animation = NewTwinkle(400, foreColour, backColour, runtimeMs)
	c.nextAnimation = nil
	c.calibrate = calibrate
	c.animationTime = animationTime

	c.runtimeMs = runtimeMs
	c.frameRate = frameRate
	c.transition = 0.0
	c.transitionTimeSecs = 5.0
	c.transitionIncrement = 1.0 / (c.frameRate * c.transitionTimeSecs)

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

func (c *Controller) cycleAnimation() {
	log.Printf("type: %v", reflect.TypeOf(c.animation).Elem().String())
	if reflect.TypeOf(c.animation).Elem().String() == "stream.GradientTrail" {
		backColour, _ := colorful.Hex("#000005") //("#100505")
		foreColour, _ := colorful.Hex("#808080") //("#404040")
		c.nextAnimation = NewTwinkle(400, foreColour, backColour, c.runtimeMs)
	} else {
		gradient := GradientTable{
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
		c.nextAnimation = NewGradientTrail(gradient, 180, 0.06, c.runtimeMs, -0.03)
	}
}

// Run causes the Controller to cycle through animations.
func (c *Controller) Run() {
	publishTimer := time.NewTicker(c.animationTime)
	for {
		select {
		case <-publishTimer.C:
			c.cycleAnimation()
		case <-c.calibrate.CalStart:
			c.animation = c.calibrate
		}
	}
}
