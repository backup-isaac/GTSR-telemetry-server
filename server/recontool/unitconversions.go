package recontool

import "gonum.org/v1/gonum/floats"

// MetersPerSecondToMilesPerHour converts a series of velocities in m/s to mph
func MetersPerSecondToMilesPerHour(metersPerSecond []float64) []float64 {
	return scaleBy(metersPerSecond, 2.234)
}

// MetersToMiles converts a series of distances in meters to miles
func MetersToMiles(meters []float64) []float64 {
	return scaleBy(meters, 0.000621371)
}

// SecondsToMinutes converts a series of times in seconds to minutes
func SecondsToMinutes(seconds []float64) []float64 {
	return scaleBy(seconds, 1.0/60)
}

func scaleBy(series []float64, k float64) []float64 {
	scaled := make([]float64, len(series))
	floats.ScaleTo(scaled, k, series)
	return scaled
}
