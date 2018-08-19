package computations

import (
	"sync"

	"telemetry-server/datatypes"
)

// TestComputation is a test computation, which is used to ensure
// the computations package is functioning properly when the
// data generator is running.
// It computes the average of the last 10 Test values
type TestComputation struct {
	values []float64
	sync.Mutex
}

// Update appends the value to the list and returns true if the list size
// is at least 10
func (tc *TestComputation) Update(point *datatypes.Datapoint) bool {
	tc.Lock()
	defer tc.Unlock()
	tc.values = append(tc.values, point.Value)
	return len(tc.values) >= 10
}

// Compute computes the average of the values tracked by the TestComputation
func (tc *TestComputation) Compute() *datatypes.Datapoint {
	tc.Lock()
	defer tc.Unlock()
	sum := float64(0)
	for _, value := range tc.values {
		sum += value
	}
	var val float64
	if len(tc.values) > 0 {
		val = sum / float64(len(tc.values))
	} else {
		val = 0
	}
	tc.values = make([]float64, 0, 10)
	return &datatypes.Datapoint{
		Metric: "Test_Computation",
		Value:  val,
	}
}

func init() {
	Register(&TestComputation{}, []string{"Test"})
}
