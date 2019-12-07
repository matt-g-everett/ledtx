package stream

// An Animation implements a way to render a specific animation.
type Animation interface {
	CalculateFrame(runtimeMs int64) *Frame
}
