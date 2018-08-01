package listener_test

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.gatech.edu/GTSR/telemetry-server/datatypes"
	"github.gatech.edu/GTSR/telemetry-server/listener"
)

func TestPacketParser(t *testing.T) {
	canConfig := map[int]*listener.CanConfigType{
		0: {
			CanID:    0,
			Datatype: "int32",
			Name:     "Test 1",
			Offset:   0,
		},
	}
	parser := listener.NewPacketParser(canConfig)
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
	expectedPoint := &datatypes.Datapoint{
		Metric: "Test 1",
		Value:  12345,
	}
	assert.Equal(t, expectedPoint, point)
}
