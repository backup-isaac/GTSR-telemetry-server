package computations_test

import (
	"telemetry-server/computations"
	"telemetry-server/datatypes"
	"telemetry-server/listener"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockComputable struct {
	values []float64
}

func (mc *mockComputable) Update(point *datatypes.Datapoint) bool {
	mc.values = append(mc.values, point.Value)
	return len(mc.values) >= 2
}

func (mc *mockComputable) Compute() *datatypes.Datapoint {
	point := &datatypes.Datapoint{
		Metric: "Result Metric",
		Value:  mc.values[0] + mc.values[1],
	}
	mc.values = make([]float64, 0, 2)
	return point
}

func (mc *mockComputable) GetMetrics() []string {
	return []string{"Computable_Integration_Test_Metric_1", "Computable_Integration_Test_Metric_2"}
}

func TestComputations(t *testing.T) {
	computations.Register(&mockComputable{})
	go computations.RunComputations()

	publisher := listener.GetDatapointPublisher()
	stream := make(chan *datatypes.Datapoint, 1000)
	err := publisher.Subscribe(stream)
	assert.NoError(t, err)

	<-time.After(100 * time.Millisecond)

	publisher.Publish(&datatypes.Datapoint{
		Metric: "Computable_Integration_Test_Metric_1",
		Value:  1,
	})
	<-stream
	select {
	case <-stream:
		assert.Fail(t, "Computable should not have triggered at this point.")
	case <-time.After(10 * time.Millisecond):
	}

	publisher.Publish(&datatypes.Datapoint{
		Metric: "Nonexistent metric",
		Value:  10,
	})
	<-stream
	select {
	case <-stream:
		assert.Fail(t, "Computable should not have triggered at this point.")
	case <-time.After(10 * time.Millisecond):
	}

	publisher.Publish(&datatypes.Datapoint{
		Metric: "Computable_Integration_Test_Metric_2",
		Value:  100,
	})
	<-stream
	var point *datatypes.Datapoint
	select {
	case point = <-stream:
	case <-time.After(time.Second):
		assert.Fail(t, "Computable should have triggered - timed out after 1 second.")
	}
	assert.Equal(t, "Result Metric", point.Metric)
	assert.Equal(t, float64(101), point.Value)
}
