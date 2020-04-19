package recontool

import (
	"math"

	"gonum.org/v1/gonum/unit/constant"
)

// Velocity computes the vehicle's linear velocity in meters per second from motor rpm and radius
func Velocity(rpm, rMot float64) float64 {
	return rpm * math.Pi * rMot / 30
}

// MotorTorque computes empirical motor torque in Nm from motor rpm and motor phase C current
func MotorTorque(rpm, iPhaseC, tMax float64) float64 {
	torque := iPhaseC * (-0.0003*rpm + 1.4292)
	if torque > tMax {
		return tMax
	}
	return torque
}

// ModeledMotorForce derives the magnitude of the force that the motors exert to cause the car to move
//       drag   rolling friction       terrain   net
// Fmot = Fd + (Crr1 + Crr2 * dxdt)Fn + Fgsinθ + ma
func ModeledMotorForce(dxdt, dvdt, theta float64, vehicle *Vehicle) float64 {
	return vehicle.DragForce(dxdt) + vehicle.RollingFrictionalForce(dxdt, theta) + vehicle.M*float64(constant.StandardGravity)*math.Sin(theta) + vehicle.M*dvdt
}

// MotorPower computes motor power from torque, velocity, and drivetrain characteristics
func MotorPower(tMot, v, iPhaseC, effDt float64, vehicle *Vehicle) float64 {
	return tMot*v/(vehicle.RMot*effDt) + 3*vehicle.RLine*iPhaseC*iPhaseC
}

// MotorEfficiency calculates motor efficiency in terms of bus voltage and torque
// should probably figure out where all these nice magic numbers come from
func MotorEfficiency(vBus, tMot float64) float64 {
	rpmMax := 7.62711864407*(vBus-79) + 600
	motorEff := rpmMax / (rpmMax + 0.1765*tMot)
	if motorEff > 1.0 {
		return 1.0
	}
	return motorEff
}

// ModelDerivedPower computes motor power from model-derived force, velocity, and drivetrain efficiency
func ModelDerivedPower(fRes, dxdt, effDt float64) float64 {
	return fRes * dxdt / effDt
}
