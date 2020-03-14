package recontool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRiemannSumIntegrate(t *testing.T) {
	assert.Equal(t, []float64{}, RiemannSumIntegrate([]float64{}, 50))
	assert.InDeltaSlice(t, []float64{
		2, 6, 0, -8, -8,
	}, RiemannSumIntegrate([]float64{
		1, 2, -3, -4, 0,
	}, 2), fd(8))
}

func TestGradient(t *testing.T) {
	assert.InDeltaSlice(t, []float64{
		0.2, 0.2,
	}, Gradient([]float64{
		-1, 1,
	}, 10), fd(0.2))
	assert.InDeltaSlice(t, []float64{
		0.2, 0.25, 0.05, 0, -0.1, -0.4,
	}, Gradient([]float64{
		-1, 1, 4, 2, 4, 0,
	}, 10), fd(0.4))
}
