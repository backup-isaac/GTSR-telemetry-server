package recontool

import "gonum.org/v1/gonum/floats"

const (
	// MetersPerSecondToMilesPerHour converts m/s -> mi/hr
	MetersPerSecondToMilesPerHour = 2.234
	// MetersToMiles converts m -> mi
	MetersToMiles = 0.000621371
	// SecondsToMinutes converts s -> min
	SecondsToMinutes = 1.0 / 60
)

// Scale scales series by k
func Scale(series []float64, k float64) []float64 {
	scaled := make([]float64, len(series))
	floats.ScaleTo(scaled, k, series)
	return scaled
}
