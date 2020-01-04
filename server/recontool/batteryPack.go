package recontool

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/floats"
)

// PackEfficiency calculates pack efficiency from the amount of power dissipated in the pack from the high voltage bus
func PackEfficiency(iBus, pBus, rPack float64) float64 {
	if pBus == 0 {
		return 1
	}
	powerRatio := iBus * iBus * rPack / math.Abs(pBus)
	return math.Abs(1 / (1 + powerRatio))
}

// PackResistanceRegression calculates battery pack resistance as -∂V/∂I of bus voltage/current, using least squares linear regression
// Returns (resistance, Y intercept of regression, bus current used, bus voltage used)
// Originally this was done as
//   IX = [ones(length(Ipack),1),Ipack];
// 	 Rmat = (IX'*IX)^-1 * IX' * Vpack;
// 	 Rpack = -Rmat(2);
//   Yintercept = Rmat(1);
// I calculate
//   sumI = ∑i for all i in Ipack
//   sumISquares = ∑i^2 for all i in Ipack
//   I_prime_resistance[i] = -sumI + length(Ipack) * Ipack[i]
//   I_prime_intercept[i] = sumISquares - sumI * Ipack[i]
//   denominator = length(Ipack)*sumISquares - sumI^2
//   Rpack = (-I_prime_resistance • Vpack) / denominator
//   Yintercept = (-I_prime_intercept • Vpack) / denominator
// The two calculations are algebraically equivalent, but mine requires less matrix math and O(1) additional space
// As previously, points with zero or impossibly low bus voltage are discarded to avoid bias
func PackResistanceRegression(busCurrent, busVoltage []float64) (float64, float64, []float64, []float64) {
	sumI := 0.0
	sumI2 := 0.0
	usedCurrents := make([]float64, 0, len(busCurrent))
	usedVoltages := make([]float64, 0, len(busVoltage))
	for i, iBus := range busCurrent {
		if busVoltage[i] < 1 {
			continue
		}
		sumI += iBus
		sumI2 += iBus * iBus
		usedCurrents = append(usedCurrents, iBus)
		usedVoltages = append(usedVoltages, busVoltage[i])
	}
	count := float64(len(usedCurrents))
	dotProductResistance := 0.0
	dotProductIntercept := 0.0
	for i, iBus := range usedCurrents {
		dotProductResistance += (-1*sumI + count*iBus) * usedVoltages[i]
		dotProductIntercept += (sumI2 - sumI*iBus) * usedVoltages[i]
	}
	denominator := count*sumI2 - (sumI * sumI)
	return (-1 * dotProductResistance / denominator), (dotProductIntercept / denominator), usedCurrents, usedVoltages
}

// PackModuleVoltages arranges raw module voltages into a 2-D array, calculates the max and min modules over time, and calculates the max-min difference over time
// Returns (raw module voltages, max-min difference, max module, min module)
func PackModuleVoltages(data map[string][]float64, vSer uint) ([][]float64, []float64, []float64, []float64) {
	rawModuleVoltages := make([][]float64, vSer)
	var i uint
	for i = 0; i < vSer; i++ {
		rawModuleVoltages[i] = data[fmt.Sprintf("Cell_Voltage_%d", i+1)]
	}
	seriesLength := len(rawModuleVoltages[0])
	maxMinDiff := make([]float64, seriesLength)
	argmaxes := make([]float64, seriesLength)
	argmins := make([]float64, seriesLength)
	for j := 0; j < seriesLength; j++ {
		var argmax uint
		var argmin uint
		for i = 0; i < vSer; i++ {
			if rawModuleVoltages[i][j] > rawModuleVoltages[argmax][j] {
				argmax = i
			} else if rawModuleVoltages[i][j] < rawModuleVoltages[argmin][j] {
				argmin = i
			}
		}
		argmaxes[j] = float64(argmax + 1)
		argmins[j] = float64(argmin + 1)
		maxMinDiff[j] = rawModuleVoltages[argmax][j] - rawModuleVoltages[argmin][j]
	}
	return rawModuleVoltages, maxMinDiff, argmaxes, argmins
}

// ModuleResistance calculates the resistance of the module as
// Rmod = (max(Vmod) - min(Vmod)) / (max(Imod) - min(Imod))
func ModuleResistance(moduleVoltage, moduleCurrent []float64) float64 {
	return (floats.Max(moduleVoltage) - floats.Min(moduleVoltage)) / (floats.Max(moduleCurrent) - floats.Min(moduleCurrent))
}

// CalculateModuleResistances returns the resistances of all battery modules
func CalculateModuleResistances(data map[string][]float64, vSer uint) []float64 {
	moduleResistances := make([]float64, vSer)
	var i uint
	for i = 0; i < vSer; i++ {
		moduleResistances[i] = ModuleResistance(data[fmt.Sprintf("Cell_Voltage_%d", i+1)], data["BMS_Current"])
	}
	return moduleResistances
}
