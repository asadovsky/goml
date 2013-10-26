// Package core defines common data structures and functions.
package core

// Note, go default-initializes array values to 0, so we can't be as fast as C.
// https://groups.google.com/d/topic/golang-nuts/Nt8js4TgB04/discussion

type Feature struct {
	Values []float32
}

func NewFeature(values []float32) *Feature {
	return &Feature{Values: values}
}

func (f *Feature) Size() int {
	return len(f.Values)
}

func (f *Feature) Value(offset int) float32 {
	return f.Values[offset]
}
