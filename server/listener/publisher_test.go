package listener

import (
	"testing"
	"time"

	"server/datatypes"

	"github.com/stretchr/testify/assert"
)

func TestDatapointPublisher(t *testing.T) {
	publisher := GetDatapointPublisher()
	defer publisher.Close()
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

	// Ensure GetDatapointPublisher does singleton logic correctly
	publisher2 := GetDatapointPublisher()
	publisher2.Publish(datapoint)
	select {
	case <-c:
	case <-time.After(time.Second):
		assert.Fail(t, "Timed out 1 second after publish")
	}
}

func TestSpecificSubscribers(t *testing.T) {
	publisher := GetDatapointPublisher()
	defer publisher.Close()
	c1 := make(chan *datatypes.Datapoint, 1)
	err := publisher.Subscribe(c1)
	assert.NoError(t, err)
	c2 := make(chan *datatypes.Datapoint, 1)
	err = publisher.Subscribe(c2, "metric1", "metric2")
	assert.NoError(t, err)

	publisher.Publish(&datatypes.Datapoint{Metric: "metric1"})
	assertReceived(t, c1)
	assertReceived(t, c2)

	publisher.Publish(&datatypes.Datapoint{Metric: "metric2"})
	assertReceived(t, c1)
	assertReceived(t, c2)

	publisher.Publish(&datatypes.Datapoint{Metric: "metric3"})
	assertReceived(t, c1)
	assertNotReceived(t, c2, "datapoint delivered to channel not subscribed to metric3")
}

func TestUnsubscribe(t *testing.T) {
	publisher := GetDatapointPublisher()
	defer publisher.Close()
	c1 := make(chan *datatypes.Datapoint, 1)
	c2 := make(chan *datatypes.Datapoint, 1)
	err := publisher.Subscribe(c1)
	assert.NoError(t, err)
	err = publisher.Subscribe(c2, "metric1")
	assert.NoError(t, err)
	publisher.Publish(&datatypes.Datapoint{Metric: "metric1"})
	assertReceived(t, c1)
	assertReceived(t, c2)
	err = publisher.Unsubscribe(c1)
	assert.NoError(t, err)
	publisher.Publish(&datatypes.Datapoint{Metric: "metric1"})
	assertNotReceived(t, c1)
	assertReceived(t, c2)
	err = publisher.Unsubscribe(c2)
	assert.NoError(t, err)
	publisher.Publish(&datatypes.Datapoint{Metric: "metric1"})
	assertNotReceived(t, c1)
	assertNotReceived(t, c2)
	close(c1)
	close(c2)
}

func assertReceived(t *testing.T, c chan *datatypes.Datapoint, message ...interface{}) {
	select {
	case <-c:
	case <-time.After(100 * time.Millisecond):
		t.Error(message...)
	}
}

func assertNotReceived(t *testing.T, c chan *datatypes.Datapoint, message ...interface{}) {
	select {
	case <-c:
		t.Error(message...)
	case <-time.After(100 * time.Millisecond):
	}
}
