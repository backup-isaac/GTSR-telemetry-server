package computations

import (
	"math"
	"server/datatypes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func computationRunner(t *testing.T, comp Computable, inputs []*datatypes.Datapoint, expectedResult *datatypes.Datapoint) {
	for i := 0; i < len(inputs)-1; i++ {
		assert.False(t, comp.Update(inputs[i]), "%T erroneously signaled an update after receiving %v\n", comp, inputs[i])
	}
	assert.True(t, comp.Update(inputs[len(inputs)-1]))
	actualResult := comp.Compute()
	assert.Equal(t, expectedResult.Metric, actualResult.Metric)
	assert.InDelta(t, expectedResult.Value, actualResult.Value, math.Abs(expectedResult.Value)/100000000.0)
	assert.Equal(t, expectedResult.Time, actualResult.Time)
	assert.Equal(t, expectedResult.Tags, actualResult.Tags)
}

var pointTime = time.Now()

func makeDatapoint(metric string, value float64) *datatypes.Datapoint {
	pointTime = pointTime.Add(time.Millisecond)
	return &datatypes.Datapoint{
		Metric: metric,
		Value:  value,
		Time:   pointTime,
	}
}
