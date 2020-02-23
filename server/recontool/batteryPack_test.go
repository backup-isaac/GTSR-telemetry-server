package recontool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateModuleResistances(t *testing.T) {
	assert.Equal(t, []float64{}, CalculateModuleResistances(map[string][]float64{}, 0))
	assert.InDeltaSlice(t, []float64{
		0.107142857, 0.071428571, 0.042857143,
	}, CalculateModuleResistances(map[string][]float64{
		"Cell_Voltage_1": {3.9, 3.8, 3.95},
		"Cell_Voltage_2": {3.92, 3.85, 3.95},
		"Cell_Voltage_3": {3.86, 3.82, 3.88},
		"BMS_Current":    {0, 1, -0.4},
		"Test 0":         {0, 0.1, 0.2},
	}, 3), 1)
}

func TestModuleResistance(t *testing.T) {
	assert.InDelta(t, 0.01, ModuleResistance([]float64{3.2, 3.17, 3.21, 3.27}, []float64{1, 0, 3, 10}), fd(0.01))
}

func TestPackModuleVoltages(t *testing.T) {
	rawModuleVoltages, maxMinDiff, maxModule, minModule := PackModuleVoltages(map[string][]float64{
		"Test 0":         {0, 0.1, 0.2},
		"Cell_Voltage_1": {3.9, 3.8, 3.95},
		"Cell_Voltage_2": {3.92, 3.85, 3.95},
		"Cell_Voltage_3": {3.86, 3.82, 3.88},
	}, 3)
	assert.Equal(t, [][]float64{
		{3.9, 3.8, 3.95},
		{3.92, 3.85, 3.95},
		{3.86, 3.82, 3.88},
	}, rawModuleVoltages)
	assert.InDeltaSlice(t, []float64{0.06, 0.05, 0.07}, maxMinDiff, fd(0.07))
	assert.Equal(t, []float64{2, 2, 1}, maxModule)
	assert.Equal(t, []float64{3, 1, 3}, minModule)
}

func TestPackResistanceUnfiltered(t *testing.T) {
	assert.InDelta(t, 1.0, PackResistanceUnfiltered([]float64{
		1, 0, 1, 2,
	}, []float64{
		99, 100, 99, 98,
	}), fd(1))
	assert.InDelta(t, 1.0, PackResistanceUnfiltered([]float64{
		1, 0, 1, 2,
	}, []float64{
		99.1, 100, 98.9, 98,
	}), fd(1))
	assert.InDelta(t, 1.0, PackResistanceUnfiltered([]float64{
		1, 0, 1, 2,
	}, []float64{
		1.1, 2, 0.9, 0,
	}), fd(1))
}

func TestPackResistanceRegression(t *testing.T) {
	resistance, yintercept, iUsed, vUsed := PackResistanceRegression([]float64{
		1, 0, 1, 2,
	}, []float64{
		99, 100, 99, 98,
	})
	assert.InDelta(t, 1.0, resistance, fd(1))
	assert.InDelta(t, 100.0, yintercept, fd(100))
	assert.Equal(t, iUsed, []float64{1, 0, 1, 2})
	assert.Equal(t, vUsed, []float64{99, 100, 99, 98})
	resistance, yintercept, iUsed, vUsed = PackResistanceRegression([]float64{
		1, 0, 1, 2,
	}, []float64{
		99.1, 100, 98.9, 98,
	})
	assert.InDelta(t, 1.0, resistance, fd(1))
	assert.InDelta(t, 100.0, yintercept, fd(100))
	assert.Equal(t, iUsed, []float64{1, 0, 1, 2})
	assert.Equal(t, vUsed, []float64{99.1, 100, 98.9, 98})
	resistance, yintercept, iUsed, vUsed = PackResistanceRegression([]float64{
		1, 0, 1, 2,
	}, []float64{
		99.1, 100, 98.9, 0,
	})
	assert.InDelta(t, 1.0, resistance, fd(1))
	assert.InDelta(t, 100.0, yintercept, fd(100))
	assert.Equal(t, iUsed, []float64{1, 0, 1})
	assert.Equal(t, vUsed, []float64{99.1, 100, 98.9})
}

func TestPackEfficiency(t *testing.T) {
	assert.Equal(t, 1.0, PackEfficiency(0, 0, 0.2))
	assert.InDelta(t, 0.9917355, PackEfficiency(10, 1200, 0.1), 0.9917355)
}
