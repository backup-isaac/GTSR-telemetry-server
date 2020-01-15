package recontool

import (
	"fmt"

	"gonum.org/v1/gonum/floats"
)

// RemoveSuspiciousZeros removes rows from the beginning of data where zero
// values are likely indicative of garbage data
// May be due to vehicle systems taking time to start properly reporting or
// recording their data
// If any module voltage is zero in a row, the row is deleted and then the
// next row is tested again. This repeats until no modules have zero voltage,
// then the function terminates
func RemoveSuspiciousZeros(data map[string][]float64, vSer uint) {
	rowsToDelete := 0
	colLen := len(data["Cell_Voltage_1"])
	var j uint
	for i := 0; i < colLen; i++ {
		rowDeleted := false
		for j = 0; j < vSer; j++ {
			voltage := data[fmt.Sprintf("Cell_Voltage_%d", j+1)][i]
			if voltage == 0 {
				rowsToDelete++
				rowDeleted = true
				break
			}
		}
		if !rowDeleted {
			break
		}
	}
	for metric, series := range data {
		data[metric] = series[rowsToDelete:]
	}
}

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
	var length int
	if len(l) > len(r) {
		length = len(l)
	} else {
		length = len(r)
	}
	avg := make([]float64, length)
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

// CalculatePower multiplies the values of i and v into the result
func CalculatePower(i, v []float64) []float64 {
	var length int
	if len(i) < len(v) {
		length = len(i)
	} else {
		length = len(v)
	}
	p := make([]float64, length)
	for j := 0; j < len(p); j++ {
		p[j] = i[j] * v[j]
	}
	return p
}

// LatLongToRelativeDisplacement converts a series of lat/long degrees to miles displaced
func LatLongToRelativeDisplacement(latLong []float64) []float64 {
	min := floats.Min(latLong)
	return CalculateSeries(func(params ...float64) float64 {
		return (params[0] - min) / LatLongDegreeToMile
	}, latLong)
}
