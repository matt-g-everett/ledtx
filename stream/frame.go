package stream

import (
	"encoding/binary"

	"github.com/lucasb-eyer/go-colorful"
)

const numPixels = 500

// Frame represents a frame of RGB pixels to display on an ledrx device.
type Frame struct {
	pixels [numPixels]colorful.Color
}

// NewFrame creates a new Frame instance.
func NewFrame() (*Frame) {
	f := new(Frame)
	return f
}

// MarshalBinary converts a Frame into binary data.
func (f *Frame) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 2, (numPixels * 3) + 2)
	binary.LittleEndian.PutUint16(data, numPixels)
	for _, p := range f.pixels {
		r, g, b := p.Clamped().RGB255()
		data = append(data, r, g, b)
	}

	return data, nil
}
