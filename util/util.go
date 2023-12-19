package util

import (
	"math/rand"

	"github.com/fogleman/ease"
)

func RandomiseSaturation(min float64, max float64) float64 {
	return rand.Float64()*(max-min) + min
}

func GenerateLut(length int) []float64 {
	increment := 1.0 / float64(length/2)
	lut := make([]float64, length)
	for i, j := 0, length-1; i < length/2; i, j = i+1, j-1 {
		value := float64(i) * increment
		lut[i] = ease.InOutQuad(value)
		lut[j] = ease.InOutQuad(value)
	}
	return lut
}
