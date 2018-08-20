package listener_test

import (
	"testing"
	"time"

	"telemetry-server/datatypes"
	"telemetry-server/listener"

	"github.com/stretchr/testify/assert"
)

func TestDatapointPublisher(t *testing.T) {
	publisher := listener.GetDatapointPublisher()
	c := make(chan *datatypes.Datapoint)
	err := publisher.Subscribe(c)
	assert.NoError(t, err)
	datapoint := &datatypes.Datapoint{
		Metric: "fake metric",
		Value:  2,
		Tags: map[string]string{
			"Hello": "World!",
		},
	}
	publisher.Publish(datapoint)
	var actualDatapoint *datatypes.Datapoint
	select {
	case actualDatapoint = <-c:
	case <-time.After(time.Second):
		assert.Fail(t, "Timed out 1 second after publish")
	}
	assert.Equal(t, datapoint, actualDatapoint)
}
