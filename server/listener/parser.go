package listener

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"time"

	"server/configs"
	"server/datatypes"
)

const (
	idle          ReceiverState = 0
	preambleRecvd ReceiverState = 1
)

// ReceiverState is the enumerated type for receiver state
type ReceiverState int

// PacketParser parses packets from a TCP stream from the car
type PacketParser struct {
	State          ReceiverState
	PacketBuffer   []byte
	PreambleBuffer []byte
	Offset         int
	CANConfigs     map[int][]*configs.CanConfigType
}

// NewPacketParser returns a new PacketParser with the standard implementation
func NewPacketParser(canConfigs map[int][]*configs.CanConfigType) *PacketParser {
	return &PacketParser{
		State:          idle,
		PacketBuffer:   make([]byte, 12),
		PreambleBuffer: make([]byte, 0),
		CANConfigs:     canConfigs,
	}
}

// payloadParsers maps datatype strings to methods to parse a value in bytes at a given offset
var payloadParsers map[string]func([]byte, int) (float64, error)

// ParseByte maintains the parser state machine, parsing one byte at a time
// It returns true when the full packet has been received
func (p *PacketParser) ParseByte(value byte) bool {
	switch p.State {
	case idle:
		// append the value into the PreambleBuffer
		p.PreambleBuffer = append(p.PreambleBuffer, value)
		if len(p.PreambleBuffer) == 2 {
			if p.PreambleBuffer[0] == 'G' && p.PreambleBuffer[1] == 'T' {
				p.State = preambleRecvd
				p.Offset = 2 // Preamble offset
				p.PreambleBuffer = make([]byte, 0)
			} else {
				// pop off the first element of preamble buffer, continue waiting.
				p.PreambleBuffer = p.PreambleBuffer[1:]
			}
		}
	case preambleRecvd:
		p.PacketBuffer[p.Offset] = value
		p.Offset++
		if p.Offset >= len(p.PacketBuffer) {
			p.State = idle
			return true
		}
	default:
		log.Println("Unrecognized packet parser state: ", p.State)
		p.State = idle
		p.PreambleBuffer = make([]byte, 0)

	}
	return false
}

// ParsePacket returns the datapoint parsed from the current packet saved within the parser
func (p *PacketParser) ParsePacket() []*datatypes.Datapoint {
	canID := int(binary.LittleEndian.Uint16(p.PacketBuffer[2:4]))
	canConfigs := p.CANConfigs[canID]
	points := make([]*datatypes.Datapoint, 0)
	for _, config := range canConfigs {
		point := &datatypes.Datapoint{
			Metric: config.Name,
			Time:   time.Now(),
		}
		converter, ok := payloadParsers[config.Datatype]
		if !ok {
			log.Println("Unrecognized datatype: " + config.Datatype)
			continue
		}
		value, err := converter(p.PacketBuffer[4:], config.Offset)
		if err != nil {
			log.Printf("Error parsing %s from CAN id 0x%x at offset %d: %s\n", config.Datatype, config.CanID, config.Offset, err)
			continue
		}
		point.Value = value
		if !config.CheckBounds || (config.MinValue <= point.Value && config.MaxValue >= point.Value) {
			points = append(points, point)
		}
	}
	return points
}

func init() {
	payloadParsers = make(map[string]func([]byte, int) (float64, error))
	payloadParsers["uint8"] = func(bytes []byte, offset int) (float64, error) {
		return float64(bytes[offset]), nil
	}
	payloadParsers["uint16"] = func(bytes []byte, offset int) (float64, error) {
		return float64(binary.LittleEndian.Uint16(bytes[offset : offset+2])), nil
	}
	payloadParsers["uint32"] = func(bytes []byte, offset int) (float64, error) {
		return float64(binary.LittleEndian.Uint32(bytes[offset : offset+4])), nil
	}
	payloadParsers["uint64"] = func(bytes []byte, offset int) (float64, error) {
		return float64(binary.LittleEndian.Uint64(bytes[offset : offset+8])), nil
	}
	payloadParsers["int8"] = func(bytes []byte, offset int) (float64, error) {
		return float64(int8(bytes[offset])), nil
	}
	payloadParsers["int16"] = func(bytes []byte, offset int) (float64, error) {
		return float64(int16(binary.LittleEndian.Uint16(bytes[offset : offset+2]))), nil
	}
	payloadParsers["int32"] = func(bytes []byte, offset int) (float64, error) {
		return float64(int32(binary.LittleEndian.Uint32(bytes[offset : offset+4]))), nil
	}
	payloadParsers["int64"] = func(bytes []byte, offset int) (float64, error) {
		return float64(int64(binary.LittleEndian.Uint64(bytes[offset : offset+8]))), nil
	}
	payloadParsers["float32"] = func(bytes []byte, offset int) (float64, error) {
		rawValue := binary.LittleEndian.Uint32(bytes[offset : offset+4])
		value := float64(math.Float32frombits(rawValue))
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return value, fmt.Errorf("invalid float value parsed from packet")
		}
		return value, nil
	}
	payloadParsers["float64"] = func(bytes []byte, offset int) (float64, error) {
		rawValue := binary.LittleEndian.Uint64(bytes[offset : offset+8])
		value := math.Float64frombits(rawValue)
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return value, fmt.Errorf("invalid float value parsed from packet")
		}
		return value, nil
	}
	payloadParsers["bit"] = func(bytes []byte, offset int) (float64, error) {
		byteOffset := offset / 8
		bitOffset := uint(offset % 8)
		if bytes[byteOffset]&(1<<bitOffset) != 0 {
			return 1, nil
		}
		return 0, nil
	}
}
