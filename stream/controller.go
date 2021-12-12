package stream

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

// Controller that manages animations.
type Controller struct {
	calibrate           *Calibrate
	animationIndex      int
	animationPlaylist   []string
	animation           Animation
	nextAnimation       Animation
	animationTime       time.Duration
	cycling             bool
	runtimeMs           int64
	frameRate           float64
	transition          float64
	transitionTimeSecs  float64
	transitionIncrement float64
	rainbowGradient     GradientTable
	rainbowStepGradient GradientTable
}

// NewController creates an instance of a Controller.
func NewController(runtimeMs int64, frameRate float64, animationTime time.Duration,
	calibrate *Calibrate) *Controller {

	c := new(Controller)

	c.rainbowGradient = GradientTable{
		{0.0, 1.0, 0.0},
		{6.0, 1.0, 0.04},   // Pink
		{87.0, 1.0, 0.14},  // Red
		{88.0, 1.0, 0.28},  // Orange
		{98.0, 1.0, 0.42},  // Yellow
		{180.0, 1.0, 0.56}, // Green
		{190.0, 1.0, 0.70}, // Turquiose
		{320.0, 1.0, 0.84}, // Blue
		{328.0, 1.0, 0.91}, // Violet
		{360.0, 1.0, 1.0},  // Pink wrap
	}

	c.rainbowStepGradient = GradientTable{
		{80.0, 1.0, 0.0},    // Red
		{80.0, 1.0, 0.167},  // Red
		{86.0, 1.0, 0.167},  // Orange
		{86.0, 1.0, 0.333},  // Orange
		{95.0, 1.0, 0.333},  // Yellow
		{95.0, 1.0, 0.5},    // Yellow
		{160.0, 1.0, 0.5},   // Green
		{160.0, 1.0, 0.666}, // Green
		{280.0, 1.0, 0.666}, // Blue
		{280.0, 1.0, 0.833}, // Blue
		{328.0, 1.0, 0.833}, // Violet
		{328.0, 1.0, 1.0},   // Violet
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
		"multi:monokai",
		"stripes:random",
		"gradient:purplegoldblue",
		"multi:purplegoldblue",
		"multi:random",
		"stripes:candycane",
		"gradient:pinkorangewhite",
		"multi:random3",
		"stripes:random",
		"multi:pinkblueturquoise",
		"gradient:rainbowstep",
		"multi:purplegoldblue",
		"rainbow:fixed",
		"multi:monokai",
		"stripes:candycane",
		"twinkle:random",
		"multi:redgreengold",
		"twinkle:blue",
		"stripes:random",
		"multi:random",
		"multi:pinksilverblue",
		"rainbow:random",
		"multi:purplegoldblue",
		"twinkle:random",
		"multi:random3",
		"stripes:candycane",
		"multi:random2",
		"twinkle:pink",
		"stripes:random",
		"multi:redgreengold",
		"multi:monokai",
		"rainbow:normal",
		"multi:random",
		"multi:redwhiteblue",
		"gradient:pinkorangewhite",
		"twinkle:random",
		"stripes:candycane",
		"multi:redgreengold",
		"twinkle:gold",
		"stripes:random",
		"multi:random2",
		"rainbow:random",
		"multi:purplegoldblue",
		"stripes:random",
		"twinkle:random",
		"multi:random3",
		"stripes:candycane",
		"multi:random",
		"multi:pinksilverblue",
		"twinkle:silver",
		"stripes:random",
		"multi:redwhiteblue",
		"rainbow:random",
		"multi:random3",
	}
	c.animationIndex = 0
	c.animation, _ = c.getAnimation()

	return c
}

// CalculateFrame calculates a frame using the current and next animation
func (c *Controller) CalculateFrame(runtimeMs int64) *Frame {
	var f *Frame
	c.runtimeMs = runtimeMs
	if c.nextAnimation != nil {
		f1 := c.animation.CalculateFrame(runtimeMs)
		f2 := c.nextAnimation.CalculateFrame(runtimeMs)
		if f1 == nil || f2 == nil {
			return f2
		}

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

func (c *Controller) getRandomSpeed(low float64, high float64) float64 {
	speed := (rand.Float64() * (high - low)) + low
	sign := rand.Float64() - 0.5
	if sign > 0 {
		return speed
	} else {
		return speed * -1.0
	}
}

func (c *Controller) createStripes(stripeColours []colorful.Color) GradientTable {
	numStripes := len(stripeColours)
	table := make(GradientTable, 0, numStripes)
	increment := 1.0 / float64(numStripes)
	for i := 0; i < numStripes; i++ {
		h, c, _ := stripeColours[i].Hcl()
		table = append(table, GradientTable{{h, c, float64(i) * increment}}...)
		table = append(table, GradientTable{{h, c, float64(i+1) * increment}}...)
	}

	return table
}

func (c *Controller) createKnownTwinkle(foreColour colorful.Color, backColour colorful.Color) Animation {
	return NewTwinkle(rand.Int31n(150)+10, foreColour, backColour, c.runtimeMs)
}

func (c *Controller) createRandomTwinkle(foreColour colorful.Color) (Animation, string) {
	randomColor := colorful.Hsl(rand.Float64()*360.0, 1.0, 0.02)
	animation := NewTwinkle(
		rand.Int31n(100)+10,
		foreColour, randomColor, c.runtimeMs)

	return animation, randomColor.Hex()
}

func (c *Controller) createFixedRainbow() Animation {
	return NewGradientTrail(c.rainbowGradient, 600, 0.06, c.runtimeMs, 0.0)
}

func (c *Controller) createKnownRainbow() Animation {
	return NewGradientTrail(c.rainbowGradient, 1200, 0.06, c.runtimeMs, -0.5)
}

func (c *Controller) createRandomRainbow() Animation {
	trailLength := rand.Int31n(970) + 30
	return NewGradientTrail(c.rainbowGradient, uint32(trailLength), 0.06, c.runtimeMs, c.getRandomSpeed(0, 0.5))
}

func (c *Controller) createGradient(gradient GradientTable, trailLength uint32, speed float64) Animation {
	return NewGradientTrail(gradient, trailLength, 0.06, c.runtimeMs, speed)
}

func (c *Controller) createGradientRandom(gradient GradientTable, trailLength uint32) Animation {
	return NewGradientTrail(gradient, trailLength, 0.06, c.runtimeMs, c.getRandomSpeed(0.2, 0.5))
}

func (c *Controller) createMultiTwinkle(backColours []colorful.Color) Animation {
	return NewMultiTwinkle(rand.Int31n(50)+20, backColours, c.runtimeMs)
}

func (c *Controller) createRandomStripes(numColours int) (Animation, string) {
	if numColours < 1 {
		numColours = rand.Intn(2) + 2
	}

	// 30% chance that one of the colours is white
	whiteIndex := rand.Intn(numColours * 3)
	trailLength := rand.Int31n(200) + 200
	stripeColours := make([]colorful.Color, numColours)
	for i := 0; i < numColours; i++ {
		if i == whiteIndex {
			stripeColours[i] = colorful.Hsl(0.0, 0.0, 0.4)
		} else {
			stripeColours[i] = colorful.Hsl(rand.Float64()*360.0, 1.0, 0.5)
		}
	}
	extraInfo := "colours: "
	extraInfo += c.SprintColours(stripeColours)

	stripeTable := c.createStripes(stripeColours)
	return NewGradientTrail(stripeTable, uint32(trailLength), 0.2, c.runtimeMs, c.getRandomSpeed(0.2, 0.3)), extraInfo
}

func (c *Controller) createRandomMultiTwinkle(numColours int) (Animation, string) {
	if numColours < 1 {
		numColours = rand.Intn(8) + 2
	}

	twinkleChance := rand.Int31n(50) + 20
	extraInfo := fmt.Sprintf("chance: 1:%d colours: ", twinkleChance)
	backColours := make([]colorful.Color, numColours)
	for i := 0; i < numColours; i++ {
		backColours[i] = colorful.Hsl(rand.Float64()*360.0, 1.0, 0.02)
	}
	extraInfo += c.SprintColours(backColours)

	return NewMultiTwinkle(twinkleChance, backColours, c.runtimeMs), extraInfo
}

func (c *Controller) SprintColour(colour colorful.Color) string {
	return fmt.Sprintf("{R: %0.3f, G: %0.3f, B: %0.3f}", colour.R, colour.G, colour.B)
}

func (c *Controller) SprintColours(colours []colorful.Color) string {
	colourCode := "[]colorful.Color{"
	for i := 0; i < len(colours); i++ {
		colourCode += c.SprintColour(colours[i])
		if i < len(colours)-1 {
			colourCode += ", "
		}
	}
	colourCode += "}"

	return colourCode
}

func (c *Controller) getAnimation() (Animation, string) {
	brightPurple := colorful.Hcl(328.0, 1.0, 0.06)
	brightPink := colorful.Color{R: 0.45, G: -0.54, B: 0.02}
	brightOrange := colorful.Color{R: 0.23, G: 0.04, B: -0.87}
	brightRed := colorful.Color{R: 0.8, G: 0.0, B: 0.00}
	brightWhite := colorful.Color{R: 0.08, G: 0.08, B: 0.08}
	brightBlue := colorful.Hcl(280.0, 1.0, 0.06)
	brightGold := colorful.Hcl(95.0, 1.0, 0.06)

	twinkleHighlight, _ := colorful.Hex("#808080")
	gold, _ := colorful.Hex("#050401")
	pink, _ := colorful.Hex("#100505")
	purple, _ := colorful.Hex("#050005")
	silver, _ := colorful.Hex("#030303")
	blue, _ := colorful.Hex("#000005")
	green, _ := colorful.Hex("#000500")
	red, _ := colorful.Hex("#050000")
	white, _ := colorful.Hex("#202020")

	extraInfo := ""
	var animation Animation
	switch c.animationPlaylist[c.animationIndex] {
	case "twinkle:blue":
		animation = c.createKnownTwinkle(twinkleHighlight, blue)
	case "twinkle:pink":
		animation = c.createKnownTwinkle(twinkleHighlight, pink)
	case "twinkle:random":
		animation, extraInfo = c.createRandomTwinkle(twinkleHighlight)
	case "twinkle:silver":
		animation = c.createKnownTwinkle(twinkleHighlight, silver)
	case "twinkle:gold":
		animation = c.createKnownTwinkle(twinkleHighlight, gold)
	case "multi:random":
		animation, extraInfo = c.createRandomMultiTwinkle(0)
	case "multi:random2":
		animation, extraInfo = c.createRandomMultiTwinkle(2)
	case "multi:random3":
		animation, extraInfo = c.createRandomMultiTwinkle(3)
	case "multi:purplegoldblue":
		animation = c.createMultiTwinkle([]colorful.Color{purple, gold, blue})
	case "multi:pinksilverblue":
		animation = c.createMultiTwinkle([]colorful.Color{pink, silver, blue})
	case "multi:redgreengold":
		animation = c.createMultiTwinkle([]colorful.Color{red, green, gold})
	case "multi:redwhiteblue":
		animation = c.createMultiTwinkle([]colorful.Color{red, white, blue})
	case "rainbow:normal":
		animation = c.createKnownRainbow()
	case "rainbow:random":
		animation = c.createRandomRainbow()
	case "rainbow:fixed":
		animation = c.createFixedRainbow()
	case "gradient:rainbowstep":
		animation = c.createGradientRandom(c.rainbowStepGradient, 1000)
	case "gradient:pinkorangewhite":
		gradient := c.createStripes([]colorful.Color{brightPink, brightOrange, brightWhite})
		animation = c.createGradient(gradient, 310, -0.3)
	case "gradient:purplegoldblue":
		gradient := c.createStripes([]colorful.Color{brightPurple, brightGold, brightBlue})
		animation = c.createGradient(gradient, 310, -0.3)
	case "stripes:candycane":
		gradient := c.createStripes([]colorful.Color{brightRed, brightWhite})
		animation = c.createGradientRandom(gradient, 320)
	case "stripes:random":
		animation, extraInfo = c.createRandomStripes(0)
	case "multi:pinkblue":
		animation = c.createMultiTwinkle([]colorful.Color{
			{R: 0.040, G: 0.000, B: 0.011},
			{R: 0.000, G: 0.024, B: 0.040},
			{R: 0.005, G: 0.000, B: 0.040}})
	case "multi:monokai":
		animation = c.createMultiTwinkle([]colorful.Color{
			{R: 0.020, G: 0.040, B: 0.000},
			{R: 0.012, G: 0.000, B: 0.040},
			{R: 0.000, G: 0.040, B: 0.019},
			{R: 0.040, G: 0.000, B: 0.005},
			{R: 0.033, G: 0.040, B: 0.000},
			{R: 0.040, G: 0.000, B: 0.010},
			{R: 0.012, G: 0.040, B: 0.000}})
	}

	if len(extraInfo) > 0 {
		log.Printf("Cycling to %s; %s", c.animationPlaylist[c.animationIndex], extraInfo)
	} else {
		log.Printf("Cycling to %s", c.animationPlaylist[c.animationIndex])
	}

	return animation, extraInfo
}

func (c *Controller) cycleAnimation() {
	if c.cycling {
		c.animationIndex++
		c.animationIndex %= len(c.animationPlaylist)
		c.nextAnimation, _ = c.getAnimation()
	}
}

// Run causes the Controller to cycle through animations.
func (c *Controller) Run() {
	publishTimer := time.NewTicker(c.animationTime)
	for {
		select {
		case <-publishTimer.C:
			c.cycleAnimation()
		case start := <-c.calibrate.C:
			if start {
				c.cycling = false
				c.animation = c.calibrate
				fmt.Println("Started displaying calibration frames...")
			} else {
				c.cycling = true
				c.cycleAnimation()
			}
		}
	}
}
