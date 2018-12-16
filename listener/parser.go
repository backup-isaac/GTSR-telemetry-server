package listener

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"telemetry-server/configs"
	"telemetry-server/datatypes"
)

// These constants are used to map between
// Valid datatypes and their string representation
const (
	Uint8Type   = "uint8"
	Uint16Type  = "uint16"
	Int32Type   = "int32"
	Float32Type = "float32"
)

const (
	idle          ReceiverState = 0
	pre1          ReceiverState = 1
	pre2          ReceiverState = 2
	pre3          ReceiverState = 3
	preambleRecvd ReceiverState = 4
)

// ReceiverState is the enumerated type for receiver state
type ReceiverState int

// PacketParser is a interface describe an object which parses incoming packets from the car
type PacketParser interface {
	ParseByte(value byte) bool
	ParsePacket() []*datatypes.Datapoint
}

// NewPacketParser returns a new PacketParser with the standard implementation
func NewPacketParser(canConfigs map[int][]*configs.CanConfigType) PacketParser {
	return &packetParser{
		State:        idle,
		PacketBuffer: make([]byte, 16),
		CANConfigs:   canConfigs,
	}
}

type packetParser struct {
	State        ReceiverState
	PacketBuffer []byte
	Offset       int
	CANConfigs   map[int][]*configs.CanConfigType
}

// ParseByte maintains the parser state machine, parsing one byte at a time
// It returns true when the full packet has been received
func (p *packetParser) ParseByte(value byte) bool {
	switch p.State {
	case idle:
		if value == byte('G') {
			p.State = pre1
		} else {
			p.State = idle
		}
	case pre1:
		if value == byte('T') {
			p.State = pre2
		} else {
			p.State = idle
		}
	case pre2:
		if value == byte('S') {
			p.State = pre3
		} else {
			p.State = idle
		}
	case pre3:
		if value == byte('R') {
			p.State = preambleRecvd
			p.Offset = 4 // Preamble offset
		} else {
			p.State = idle
		}
	case preambleRecvd:
		p.PacketBuffer[p.Offset] = value
		p.Offset++
		if p.Offset >= len(p.PacketBuffer) {
			p.State = idle
			return true
		}
	default:
		fmt.Println("Unrecognized packet parser state: ", p.State)
		p.State = idle
	}
	return false
}

// ParsePacket returns the datapoint parsed from the current packet saved within the parser
func (p *packetParser) ParsePacket() []*datatypes.Datapoint {
	canID := int(binary.LittleEndian.Uint16(p.PacketBuffer[4:6]))
	canConfigs := p.CANConfigs[canID]
	points := make([]*datatypes.Datapoint, 0)
	for _, config := range canConfigs {
		point := &datatypes.Datapoint{
			Metric: config.Name,
			Time:   time.Now(),
		}
		if config.Datatype == Uint8Type {
			point.Value = float64(p.PacketBuffer[8+config.Offset])
		} else if config.Datatype == Uint16Type {
			point.Value = float64(binary.LittleEndian.Uint16(p.PacketBuffer[8+config.Offset : 10+config.Offset]))
		} else if config.Datatype == Int32Type {
			point.Value = float64(binary.LittleEndian.Uint32(p.PacketBuffer[8+config.Offset : 12+config.Offset]))
		} else if config.Datatype == Float32Type {
			rawValue := binary.LittleEndian.Uint32(p.PacketBuffer[8+config.Offset : 12+config.Offset])
			point.Value = float64(math.Float32frombits(rawValue))
			if math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
				continue
			}
		} else {
			fmt.Println("Unrecognized datatype: " + config.Datatype)
			continue
		}
		if !config.CheckBounds || (config.MinValue <= point.Value && config.MaxValue >= point.Value) {
			points = append(points, point)
		}
	}
	return points
}
