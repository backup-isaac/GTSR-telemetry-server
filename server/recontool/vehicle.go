package recontool

import (
	"math"

	"gonum.org/v1/gonum/unit/constant"
)

// Vehicle contains parameters of the vehicle whose data is being analyzed
type Vehicle struct {
	// Motor radius (m)
	RMot float64
	// Mass (kg)
	M float64
	// Crr1 rolling resistance
	Crr1 float64
	// Crr2 dynamic rolling resistance (s/m)
	Crr2 float64
	// Area drag coefficient (m^2)
	CDa float64
	// Maximum motor torque (N-m)
	TMax float64
	// Battery charge capacity (A-hr)
	QMax float64
	// Phase line resistance (Ω)
	RLine float64
	// Maximum battery module voltage (V)
	VcMax float64
	// Minimum battery module voltage (V)
	VcMin float64
	// Number of battery modules in series
	VSer uint
}

// density of air
const rho = 1.225

// DragForce computes aerodynamic drag experienced by this vehicle at the given velocity
func (v *Vehicle) DragForce(velocity float64) float64 {
	return 0.5 * rho * v.CDa * velocity * velocity
}

// RollingFrictionalForce computes the rolling frictional force experienced by this vehicle at the given velocity and angle wrt horizontal
func (v *Vehicle) RollingFrictionalForce(velocity, theta float64) float64 {
	coefficients := v.Crr1*math.Cos(theta) + v.Crr2*velocity
	return v.M * float64(constant.StandardGravity) * coefficients
}
