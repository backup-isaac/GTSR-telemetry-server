package listener_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.gatech.edu/JDuncan45/telemetry-server/listener"
)

func TestDatapointPublisher(t *testing.T) {
	listener.Init()
	c := make(chan *listener.Datapoint)
	err := listener.Subscribe(c)
	assert.NoError(t, err)
	datapoint := &listener.Datapoint{
		Metric: "fake metric",
		Value:  2,
		Tags: map[string]string{
			"Hello": "World!",
		},
	}
	listener.Publish(datapoint)
	var actualDatapoint *listener.Datapoint
	select {
	case actualDatapoint = <-c:
	case <-time.After(time.Second):
		assert.Fail(t, "Timed out 1 second after publish")
	}
	assert.Equal(t, datapoint, actualDatapoint)
}
