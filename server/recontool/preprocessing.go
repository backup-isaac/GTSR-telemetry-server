package recontool

import (
	"math"

	"gonum.org/v1/gonum/floats"
)

// RemoveTimeOffsets makes t linearly spaced and subtracts
// t[0] from every value
func RemoveTimeOffsets(t []int64) []float64 {
	end := float64(t[len(t)-1])
	start := float64(t[0])
	alteredT := make([]float64, len(t))
	return floats.Span(alteredT, 0, (end-start)/1000)
}

// Average averages the values of l and of r into the result
// If one of the arguments (say l) is longer than the other,
// result[i] where i >= len(r) = l[i]
func Average(l, r []float64) []float64 {
	avg := make([]float64, int(math.Max(float64(len(l)), float64(len(r)))))
	for i := 0; i < len(avg); i++ {
		if i >= len(l) {
			avg[i] = r[i]
		} else if i >= len(r) {
			avg[i] = l[i]
		} else {
			avg[i] = (l[i] + r[i]) / 2
		}
	}
	return avg
}
