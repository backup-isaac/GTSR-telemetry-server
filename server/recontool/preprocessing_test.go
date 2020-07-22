package recontool

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatLongToRelativeDisplacement(t *testing.T) {
	assert.Equal(t, []float64{0}, LatLongToRelativeDisplacement([]float64{150}))
	assert.InDeltaSlice(t, []float64{2 * 69, 1 * 69, 0, 3 * 69}, LatLongToRelativeDisplacement([]float64{1, 0, -1, 2}), fd(3*69))
}

func TestCalculatePower(t *testing.T) {
	assert.Equal(t, []float64{}, CalculatePower([]float64{}, []float64{}))
	assert.InDeltaSlice(t, []float64{
		100, 202, 306,
	}, CalculatePower([]float64{
		1, 2, 3,
	}, []float64{
		100, 101, 102,
	}), fd(306))
	assert.InDeltaSlice(t, []float64{
		100, 202, 306,
	}, CalculatePower([]float64{
		1, 2, 3, 4,
	}, []float64{
		100, 101, 102,
	}), fd(306))
	assert.InDeltaSlice(t, []float64{
		100, 202, 306,
	}, CalculatePower([]float64{
		1, 2, 3,
	}, []float64{
		100, 101, 102, 103,
	}), fd(306))
}

func TestAverage(t *testing.T) {
	assert.Equal(t, []float64{}, Average([]float64{}, []float64{}))
	assert.InDeltaSlice(t, []float64{
		2, 0.5, 0,
	}, Average([]float64{
		1, 2, -3,
	}, []float64{
		3, -1, 3,
	}), fd(2))
	assert.InDeltaSlice(t, []float64{
		2, 0.5, 0,
	}, Average([]float64{
		1, 2, -3, -4,
	}, []float64{
		3, -1, 3,
	}), fd(2))
	assert.InDeltaSlice(t, []float64{
		2, 0.5, 0,
	}, Average([]float64{
		1, 2, -3,
	}, []float64{
		3, -1, 3, 0,
	}), fd(2))
}

func TestRemoveTimeOffsets(t *testing.T) {
	assert.InDeltaSlice(t, []float64{0, 0.001}, RemoveTimeOffsets([]int64{900, 901}), fd(0.001))
	assert.InDeltaSlice(t, []float64{0, 1, 2}, RemoveTimeOffsets([]int64{-1000, 0, 1000}), fd(2))
	assert.InDeltaSlice(t, []float64{0, 3.75, 7.5, 11.25, 15}, RemoveTimeOffsets([]int64{1000, 2000, 4000, 8000, 16000}), fd(15))
}

func TestRemoveSuspiciousZeros(t *testing.T) {
	testSuspiciousZerosRunner(t, []int{})
	testSuspiciousZerosRunner(t, []int{4})
	testSuspiciousZerosRunner(t, []int{4, 4, 4})
	testSuspiciousZerosRunner(t, []int{4, 2, 4})
}

func testSuspiciousZerosRunner(t *testing.T, suspiciousZeros []int) {
	inputData, inputTimestamps, expectedData, expectedTimestamps := createSuspiciousData(suspiciousZeros)
	actualTimestamps := RemoveSuspiciousZeros(inputData, inputTimestamps, 4)
	assert.Equal(t, expectedTimestamps, actualTimestamps)
	assert.Equal(t, expectedData, inputData)
}

func createSuspiciousData(suspiciousZeros []int) (map[string][]float64, []int64, map[string][]float64, []int64) {
	colLen := len(suspiciousZeros) + 5
	inputData := map[string][]float64{
		"Test 0": make([]float64, colLen),
	}
	inputTimestamps := make([]int64, colLen)
	expectedData := map[string][]float64{
		"Test 0": make([]float64, 5),
	}
	expectedTimestamps := make([]int64, 5)
	for i := 1; i <= 4; i++ {
		colName := fmt.Sprintf("Cell_Voltage_%d", i)
		inputData[colName] = make([]float64, colLen)
		expectedData[colName] = make([]float64, 5)
	}
	for i := 0; i < colLen; i++ {
		inputData["Test 0"][i] = float64(i * i)
		for j := 1; j <= 4; j++ {
			colName := fmt.Sprintf("Cell_Voltage_%d", j)
			if i < len(suspiciousZeros) && suspiciousZeros[i] == j {
				inputData[colName][i] = 0
			} else if i == colLen-1 && j%2 == 0 {
				inputData[colName][i] = 0
			} else {
				inputData[colName][i] = float64(i + j)
			}
		}
		inputTimestamps[i] = int64(i) * 100
		if i >= len(suspiciousZeros) {
			expectedData["Test 0"][i-len(suspiciousZeros)] = float64(i * i)
			expectedTimestamps[i-len(suspiciousZeros)] = int64(i) * 100
			for j := 1; j <= 4; j++ {
				colName := fmt.Sprintf("Cell_Voltage_%d", j)
				if i == colLen-1 && j%2 == 0 {
					expectedData[colName][i-len(suspiciousZeros)] = 0
				} else {
					expectedData[colName][i-len(suspiciousZeros)] = float64(i + j)
				}
			}
		}
	}
	return inputData, inputTimestamps, expectedData, expectedTimestamps
}
