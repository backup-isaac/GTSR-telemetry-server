package recontool

import "math"

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

// CalculateMotorTorque computes motor torque in Nm from motor rpm and motor phase C current
func CalculateMotorTorque(rpm, iPhaseC float64) float64 {
	return iPhaseC * (-0.0003*rpm + 1.4292)
}

// CalculateMotorTorqueSeries computes motor torque for a series of RPM and phase C current points
func CalculateMotorTorqueSeries(rpm []float64, iPhaseC []float64) []float64 {
	motorTorqueSeries := make([]float64, len(rpm))
	for i, r := range rpm {
		motorTorqueSeries[i] = CalculateMotorTorque(r, iPhaseC[i])
	}
	return motorTorqueSeries
}
