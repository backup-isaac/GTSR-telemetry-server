package recontool

import (
	"math"

	"gonum.org/v1/gonum/unit/constant"
)

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
func CalculateMotorTorqueSeries(rpm, iPhaseC []float64, maxTorque float64) []float64 {
	motorTorqueSeries := make([]float64, len(rpm))
	for i, r := range rpm {
		motorTorqueSeries[i] = CalculateMotorTorque(r, iPhaseC[i])
		if motorTorqueSeries[i] > maxTorque {
			motorTorqueSeries[i] = maxTorque
		}
	}
	return motorTorqueSeries
}

// DeriveMotorForce derives the magnitude of the force that the motors exert to cause the car to move
//       drag   rolling friction       terrain   net
// Fmot = Fd + (Crr1 + Crr2 * dxdt)Fn + Fgsinθ + ma
func DeriveMotorForce(dxdt, dvdt, theta float64, vehicle *Vehicle) float64 {
	return vehicle.DragForce(dxdt) + vehicle.RollingFrictionalForce(dxdt, theta) + vehicle.M*float64(constant.StandardGravity)*math.Sin(theta) + vehicle.M*dvdt
}

// DeriveMotorForceSeries computes model-derived motor force for a series of points
func DeriveMotorForceSeries(dxdt, dvdt, theta []float64, vehicle *Vehicle) []float64 {
	motorForceSeries := make([]float64, len(theta))
	for i, dx := range dxdt {
		motorForceSeries[i] = DeriveMotorForce(dx, dvdt[i], theta[i], vehicle)
	}
	return motorForceSeries
}

// CalculateMotorPower computes motor power from torque, velocity, and drivetrain characteristics
func CalculateMotorPower(tMot, v, iPhaseC, effDt float64, vehicle *Vehicle) float64 {
	return tMot*v/(vehicle.RMot*effDt) + 3*vehicle.RLine*iPhaseC*iPhaseC
}

// CalculateMotorPowerSeries computes motor power for a series of points
func CalculateMotorPowerSeries(tMot, dxdt, iPhaseC, effDt []float64, vehicle *Vehicle) []float64 {
	motorPowerSeries := make([]float64, len(tMot))
	for i, t := range tMot {
		motorPowerSeries[i] = CalculateMotorPower(t, dxdt[i], iPhaseC[i], effDt[i], vehicle)
	}
	return motorPowerSeries
}

// CalculateMotorEfficiency calculates motor efficiency in terms of bus voltage and torque
// should probably figure out where all these nice magic numbers come from
func CalculateMotorEfficiency(vBus, tMot float64) float64 {
	rpmMax := 7.62711864407*(vBus-79) + 600
	return rpmMax / (rpmMax + 0.1765*tMot)
}

// CalculateMotorEfficiencySeries calculates motor efficiency for a series of points
func CalculateMotorEfficiencySeries(vBus, tMot []float64) []float64 {
	effMot := make([]float64, len(vBus))
	for i, v := range vBus {
		effMot[i] = CalculateMotorEfficiency(v, tMot[i])
		if effMot[i] > 1 {
			effMot[i] = 1
		}
	}
	return effMot
}
