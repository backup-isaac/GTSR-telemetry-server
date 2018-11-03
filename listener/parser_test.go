package listener_test

import (
	"encoding/binary"
	"math"
	"testing"
	"time"

	"telemetry-server/configs"
	"telemetry-server/datatypes"
	"telemetry-server/listener"

	"github.com/stretchr/testify/assert"
)

func TestPacketParser(t *testing.T) {
	canconfig := map[int][]*configs.CanConfigType{
		0: {
			{
				CanID:       0,
				Datatype:    "int32",
				Name:        "Test 1",
				Offset:      0,
				CheckBounds: false,
			},
			{
				CanID:       0,
				Datatype:    "float32",
				Name:        "Test 2",
				Offset:      4,
				CheckBounds: false,
			},
		},
	}
	parser := listener.NewPacketParser(canconfig)
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
	binary.LittleEndian.PutUint32(valueBytes, math.Float32bits(2.5))
	for i := 0; i < 4; i++ {
		packet[8+i] = valueBytes[i]
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
	points := parser.ParsePacket()
	expectedPoints := []*datatypes.Datapoint{
		{
			Metric: "Test 1",
			Value:  12345,
		},
		{
			Metric: "Test 2",
			Value:  2.5,
		},
	}
	for _, point := range points {
		point.Time = time.Time{}
	}
	assert.ElementsMatch(t, expectedPoints, points)
}

func TestInvalidValues(t *testing.T) {
	canConfig := map[int][]*configs.CanConfigType{
		0: {
			{
				CanID:       0,
				Datatype:    "int32",
				Name:        "Test 1",
				Offset:      0,
				CheckBounds: true,
				MinValue:    0,
				MaxValue:    1,
			},
			{
				CanID:       0,
				Datatype:    "float32",
				Name:        "Test 2",
				Offset:      4,
				CheckBounds: true,
				MinValue:    0.0,
				MaxValue:    1.0,
			},
		},
	}
	parser := listener.NewPacketParser(canConfig)
	packet := append([]byte("GTSR"), make([]byte, 12)...)
	packet[8] = 12
	valueBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueBytes, math.Float32bits(3.14159))
	for i := 0; i < 4; i++ {
		packet[12+i] = valueBytes[i]
	}
	for i := 0; i < len(packet); i++ {
		parser.ParseByte(packet[i])
	}
	points := parser.ParsePacket()
	assert.Equal(t, 0, len(points))
}
