package recontool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeanIf(t *testing.T) {
	assert.Equal(t, 0.0, meanIf([]float64{}, truePredicate))
	assert.Equal(t, 0.0, meanIf([]float64{1, 2, 3}, falsePredicate))
	assert.Equal(t, 3.0, meanIf([]float64{1, 2, 3, 4, 5}, truePredicate))
	assert.Equal(t, 2.5, meanIf([]float64{1, 2, 3, 4, 5, 6, 7}, func(f float64) bool {
		return f < 5
	}))
}

func truePredicate(float64) bool {
	return true
}

func falsePredicate(float64) bool {
	return false
}

func TestCalculateSeries(t *testing.T) {
	assert.Equal(t, []float64{1, 2, 3, 4, 5}, CalculateSeries(func(params ...float64) float64 {
		return params[0]
	}, []float64{1, 2, 3, 4, 5}))
	assert.Equal(t, []float64{6, 24, 60, 120, 210}, CalculateSeries(func(params ...float64) float64 {
		return params[0] * params[1] * params[2]
	}, []float64{1, 2, 3, 4, 5}, []float64{2, 3, 4, 5, 6}, []float64{3, 4, 5, 6, 7}))
}
