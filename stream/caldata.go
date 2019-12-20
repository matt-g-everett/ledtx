package stream

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type CalibrationMessage struct {
	Type string `json:"type"`
	Locations []float64 `json:"locations"`
}

type RawCalibrationData struct {
	Pixels []int32 `json:"pixels"`
	Locations []Point `json:"locations"`
}

type Bin struct {
	Location Point `json:"loc"`
	Hits int32 `json:"hits"`
	Pixels []int32 `json:"pixels"`
}

type AggregatedData struct {
	Bins []*Bin `json:"bins"`
}

type Pixel struct {
	Resolved bool `json:"resolved"`
	Location Point `json:"loc"`
}

/*type BinFrequency struct {
	Count int32 `json:"count"`
	Location Point `json:"loc"`
}

type PixelData struct {
	Pixel int32 `json:"pixel"`
	MinCount int32 `json:"minCount"`
	MaxCount int32 `json:"maxCount"`
	Bins []BinFrequency `json:"bins"`
}*/