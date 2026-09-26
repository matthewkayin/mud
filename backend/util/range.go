package util

import "math/rand/v2"

type Int32Range struct {
	Min int32
	Max int32
}

func (int32range *Int32Range) Length() int32 {
	return int32range.Max - int32range.Min
}

// Returns a random number within the range
func (int32range *Int32Range) ChooseRandom() int32 {
	if int32range.Max - int32range.Min <= 0 {
		return int32range.Min
	}
	return int32range.Min + rand.Int32N(int32range.Length())
}
