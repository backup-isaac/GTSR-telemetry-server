package recontool

import (
	"math"
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
//   denominator = length(Ipack)*sumISquares + sumI^2
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
