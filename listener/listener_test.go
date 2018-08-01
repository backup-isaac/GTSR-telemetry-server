package listener_test

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.gatech.edu/JDuncan45/telemetry-server/listener"
)

func TestDatapointPublisher(t *testing.T) {
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

func TestPacketParser(t *testing.T) {
	parser := &listener.Parser{
		PacketBuffer: make([]byte, 16),
	}
	packet := make([]byte, 12)
	// Can ID 0
	packet[0] = 0
	packet[1] = 0
	value := uint32(12345)
	valueBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueBytes, value)
	for i := 0; i < 4; i++ {
		packet[4+i] = valueBytes[i]
	}
	bytes := append([]byte("GTSR"), packet...)
	for i := 0; i < len(bytes); i++ {
		ok := parser.ParseByte(bytes[i])
		if i == len(bytes)-1 {
			assert.True(t, ok)
		} else {
			assert.False(t, ok)
		}
	}
	point := parser.ParsePacket()
	expectedPoint := &listener.Datapoint{
		Metric: "Test 1",
		Value:  12345,
	}
	assert.Equal(t, expectedPoint, point)
}
