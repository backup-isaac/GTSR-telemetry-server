package recontool

import (
	"fmt"
)

// Gradient emulates functionality of matlab's gradient(x, dt) with uniform dt
// Panics if len(x) < 2
func Gradient(x []float64, dt float64) []float64 {
	if len(x) < 2 {
		panic(fmt.Sprintf("len(x) must be >= 2, was %d", len(x)))
	}
	dxdt := make([]float64, len(x))
	dxdt[0] = (x[1] - x[0]) / dt
	for i := 1; i < len(dxdt)-1; i++ {
		dxdt[i] = (x[i+1] - x[i-1]) / (2 * dt)
	}
	dxdt[len(dxdt)-1] = (x[len(dxdt)-1] - x[len(dxdt)-2]) / dt
	return dxdt
}

// RiemannSumIntegrate integrates dxdt using a right Riemann sum
func RiemannSumIntegrate(dxdt []float64, dt float64) []float64 {
	x := make([]float64, len(dxdt))
	cumsum := 0.0
	for i := 0; i < len(x); i++ {
		cumsum += dxdt[i] * dt
		x[i] = cumsum
	}
	return x
}
