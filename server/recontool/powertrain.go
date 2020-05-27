package recontool

import (
	"math"
)

// DrivetrainEfficiency calculates total drivetrain efficiency
func DrivetrainEfficiency(mcEfficiency, packEfficiency, motorEfficiency float64) float64 {
	return mcEfficiency * packEfficiency * motorEfficiency
}

// MotorControllerEfficiency calculates motor controller efficiency
func MotorControllerEfficiency(iPhaseC, vBus, pBus float64) float64 {
	powerLoss := 0.0108*iPhaseC*iPhaseC + 3.3345e-3*math.Abs(iPhaseC) + 0.018153 + 1.5625e-4*vBus
	effMc := 1 - powerLoss/pBus
	if effMc > 1.0 {
		return 1.0
	}
	return effMc
}

// BusPower calculates high voltage bus power
func BusPower(vBusLeft, vBusRight, iBusLeft, iBusRight float64) float64 {
	return iBusLeft*vBusLeft + iBusRight*vBusRight
}
