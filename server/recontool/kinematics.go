package recontool

import (
	"math"
)

// CalculateAccelerationSeries computes the vehicle's linear acceleration in m/s^2 from velocity
// Assumes uniform dt since that's what we are working with
// otherwise emulates functionality of matlab's gradient(x, dt)
func CalculateAccelerationSeries(velocitySeries []float64, dt float64) []float64 {
	accelerationSeries := make([]float64, len(velocitySeries))
	accelerationSeries[0] = (velocitySeries[1] - velocitySeries[0]) / dt
	for i := 1; i < len(accelerationSeries)-1; i++ {
		accelerationSeries[i] = (velocitySeries[i+1] - velocitySeries[i-1]) / (2 * dt)
	}
	accelerationSeries[len(accelerationSeries)-1] = (velocitySeries[len(accelerationSeries)-1] - velocitySeries[len(accelerationSeries)-2]) / dt
	return accelerationSeries
}

// CalculateVelocity computes the vehicle's linear velocity in meters per second from motor rpm and radius
func CalculateVelocity(rpm, rMot float64) float64 {
	return rpm * math.Pi * rMot / 30
}

// CalculateVelocitySeries computes velocity for a series of RPM points
func CalculateVelocitySeries(rpm []float64, rMot float64) []float64 {
	velocitySeries := make([]float64, len(rpm))
	for i, r := range rpm {
		velocitySeries[i] = CalculateVelocity(r, rMot)
	}
	return velocitySeries
}

// CalculateDistanceSeries integrates velocity to compute distance
func CalculateDistanceSeries(velocitySeries []float64, dt float64) []float64 {
	distanceSeries := make([]float64, len(velocitySeries))
	cumsum := 0.0
	for i := 0; i < len(distanceSeries); i++ {
		cumsum += velocitySeries[i] * dt
		distanceSeries[i] = cumsum
	}
	return distanceSeries
}
