package util

import "math/rand"

func RandomiseSaturation(min float64, max float64) float64 {
	return rand.Float64()*(max-min) + min
}
