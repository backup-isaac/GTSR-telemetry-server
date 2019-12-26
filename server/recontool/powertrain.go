package recontool

import (
	"math"
)

// CalculateDrivetrainEfficiency calculates total drivetrain efficiency
func CalculateDrivetrainEfficiency(mcEfficiency, packEfficiency, motorEfficiency float64) float64 {
	return mcEfficiency * packEfficiency * motorEfficiency
}

// CalculateDrivetrainEfficiencySeries calculates drivetrain efficiency for a series of points
func CalculateDrivetrainEfficiencySeries(mcEfficiency, packEfficiency, motorEfficiency []float64) []float64 {
	dtEfficiency := make([]float64, len(mcEfficiency))
	for i, e := range mcEfficiency {
		dtEfficiency[i] = CalculateDrivetrainEfficiency(e, packEfficiency[i], motorEfficiency[i])
	}
	return dtEfficiency
}

// CalculateMotorControllerEfficiency calculates motor controller efficiency
func CalculateMotorControllerEfficiency(iPhaseC, vBus, pBus float64) float64 {
	powerLoss := 0.0108*iPhaseC*iPhaseC + 3.3345e-3*math.Abs(iPhaseC) + 0.018153 + 1.5625e-4*vBus
	return 1 - powerLoss/pBus
}

// CalculateMotorControllerEfficiencySeries calculates motor controller efficiency for a series of points
func CalculateMotorControllerEfficiencySeries(phaseCurrent, busVoltage, busPower []float64) []float64 {
	meanPbus := meanIf(busPower, func(p float64) bool { return p > 0 })
	mcEfficiency := make([]float64, len(phaseCurrent))
	for i, ip := range phaseCurrent {
		mcEfficiency[i] = CalculateMotorControllerEfficiency(ip, busVoltage[i], meanPbus)
		if mcEfficiency[i] > 1.0 {
			mcEfficiency[i] = 1.0
		}
	}
	return mcEfficiency
}

// CalculateBusPower calculates high voltage bus power
func CalculateBusPower(vBusLeft, vBusRight, iBusLeft, iBusRight float64) float64 {
	return iBusLeft*vBusLeft + iBusRight*vBusRight
}

// CalculateBusPowerSeries calculates high voltage bus power for a series of points
func CalculateBusPowerSeries(leftBusVoltage, rightBusVoltage, leftBusCurrent, rightBusCurrent []float64) []float64 {
	pBus := make([]float64, len(leftBusVoltage))
	for i, v := range leftBusVoltage {
		pBus[i] = CalculateBusPower(v, rightBusVoltage[i], leftBusCurrent[i], rightBusCurrent[i])
	}
	return pBus
}
