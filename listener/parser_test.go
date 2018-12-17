package listener_test

import (
	"encoding/binary"
	"math"
	"testing"
	"time"

	"fmt"
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
				Datatype:    "uint32",
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
				Datatype:    "uint32",
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

func TestParseConfigs(t *testing.T) {
	configsMap, err := configs.LoadConfigs() // test loading the JSON file
	assert.NoError(t, err)
	for canID, configTypes := range configsMap {
		for _, config := range configTypes {
			assert.Equal(t, canID, config.CanID,
				fmt.Sprintf("Config %+v \nhas CanID  that does not match CanID in canConfigs", *config))
			assert.True(t, config.CanID >= 0,
				fmt.Sprintf("Config %+v \nhas CanID less than 0 : %d", *config, config.CanID))
			assert.True(t, config.Offset >= 0,
				fmt.Sprintf("Config %+v \nhas offset less than 0: %d", *config, config.Offset))
			assert.True(t, config.Offset <= 7,
				fmt.Sprintf("Config %+v \nhas offset greater than 7: %d", *config, config.Offset))
			_, ok := listener.PayloadParsers[config.Datatype]
			assert.True(t, ok, fmt.Sprintf("Config: %+v \nhas an invalid datatype: %s", *config, config.Datatype))
			if config.Datatype == "int16" || config.Datatype == "uint16" {
				assert.True(t, config.Offset <= 6,
					fmt.Sprintf("Config %+v \nhas offset extending past length of a CAN payload: %d", *config, config.Offset))
			} else if config.Datatype == "float32" ||
				config.Datatype == "int32" || config.Datatype == "uint32" {
				assert.True(t, config.Offset <= 4,
					fmt.Sprintf("Config %+v \nhas offset extending past length of a CAN payload: %d", *config, config.Offset))
			} else if config.Datatype == "float64" ||
				config.Datatype == "int64" || config.Datatype == "uint64" {
				assert.True(t, config.Offset == 0,
					fmt.Sprintf("Config %+v \nhas offset extending past length of a CAN payload: %d", *config, config.Offset))
			}
		}
	}
}
