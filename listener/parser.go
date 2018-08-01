package listener

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
)

// CanConfigType holds CAN configuration information
type CanConfigType struct {
	CanID    int
	Datatype string
	Name     string
	Offset   int
}

// LoadConfigs loads the CAN configs from the config file
func LoadConfigs() (map[int]*CanConfigType, error) {
	rawJSON, err := ioutil.ReadFile("can_config.json")
	if err != nil {
		return nil, err
	}
	var canConfigList []CanConfigType
	err = json.Unmarshal(rawJSON, &canConfigList)
	if err != nil {
		return nil, err
	}
	canDatatypes := make(map[int]*CanConfigType)
	for i := range canConfigList {
		config := &canConfigList[i]
		canDatatypes[config.CanID] = config
	}
	return canDatatypes, nil
}

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
	ParsePacket() *Datapoint
}

// NewPacketParser returns a new PacketParser with the standard implementation
func NewPacketParser(canConfigs map[int]*CanConfigType) PacketParser {
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
	CANConfigs   map[int]*CanConfigType
}

const (
	uint8Type   = "uint8"
	int32Type   = "int32"
	float32Type = "float32"
)

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
func (p *packetParser) ParsePacket() *Datapoint {
	canID := int(binary.LittleEndian.Uint16(p.PacketBuffer[4:6]))
	config := p.CANConfigs[canID]
	point := &Datapoint{
		Metric: config.Name,
	}
	if config.Datatype == uint8Type {
		point.Value = p.PacketBuffer[8+config.Offset]
	} else if config.Datatype == int32Type {
		point.Value = int(binary.LittleEndian.Uint32(p.PacketBuffer[8+config.Offset : 12+config.Offset]))
	} else if config.Datatype == float32Type {
		rawValue := binary.LittleEndian.Uint32(p.PacketBuffer[8+config.Offset : 12+config.Offset])
		point.Value = math.Float32frombits(rawValue)
	}
	return point
}
