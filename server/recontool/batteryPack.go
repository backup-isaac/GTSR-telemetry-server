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

// PackResistance calculates battery pack resistance as ∂V/∂I of bus voltage/current, using least squares
// Originally this was done as
//   IX = [ones(length(Ipack),1),Ipack];
// 	 Rmat = (IX'*IX)^-1 * IX' * Vpack;
// 	 Rpack = -Rmat(2);
// I calculate
//   sumI = ∑i for all i in Ipack
//   sumISquares = ∑i^2 for all i in Ipack
//   I_prime[i] = -sumI + length(Ipack) * Ipack[i]
//   Rpack = (-I_prime * Vpack) / (length(Ipack)*sumISquares + sumI^2)
// The two calculations are algebraically equivalent, but mine requires less matrix math and O(1) additional space
func PackResistance(busCurrent, busVoltage []float64) float64 {
	sumI := 0.0
	sumI2 := 0.0
	for _, iBus := range busCurrent {
		sumI += iBus
		sumI2 += iBus * iBus
	}
	dotProduct := 0.0
	for i, iBus := range busCurrent {
		dotProduct += (-1*sumI + float64(len(busCurrent))*iBus) * busVoltage[i]
	}
	return dotProduct / (float64(len(busCurrent))*sumI2 + (sumI * sumI))
}
