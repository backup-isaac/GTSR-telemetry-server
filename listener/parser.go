package listener

import (
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"math"
)

type canConfigType struct {
	CanID    int
	Datatype string
	Name     string
	Offset   int
}

var canDatatypes map[int]*canConfigType

func loadConfigs() error {
	canDatatypes = make(map[int]*canConfigType)
	rawJSON, err := ioutil.ReadFile("../can_config.json")
	if err != nil {
		return err
	}
	var canConfigList []canConfigType
	err = json.Unmarshal(rawJSON, &canConfigList)
	if err != nil {
		return err
	}
	for i := range canConfigList {
		config := &canConfigList[i]
		canDatatypes[config.CanID] = config
	}
	return nil
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

// Parser is a struct which parses incoming packets from the car
type Parser struct {
	State        ReceiverState
	PacketBuffer []byte
	Offset       int
}

const (
	uint8Type   = "uint8"
	int32Type   = "int32"
	float32Type = "float32"
)

// ParseByte maintains the parser state machine, parsing one byte at a time
func (p *Parser) ParseByte(value byte) bool {
	if p.State == idle {
		if value == byte('G') {
			p.State = pre1
		} else {
			p.State = idle
		}
	} else if p.State == pre1 {
		if value == byte('T') {
			p.State = pre2
		} else {
			p.State = idle
		}
	} else if p.State == pre2 {
		if value == byte('S') {
			p.State = pre3
		} else {
			p.State = idle
		}
	} else if p.State == pre3 {
		if value == byte('R') {
			p.State = preambleRecvd
			p.Offset = 4
		} else {
			p.State = idle
		}
	} else if p.State == preambleRecvd {
		p.PacketBuffer[p.Offset] = value
		p.Offset++
		if p.Offset >= len(p.PacketBuffer) {
			p.State = idle
			return true
		}
	}
	return false
}

// ParsePacket returns the datapoint parsed from the current packet saved within the parser
func (p *Parser) ParsePacket() *Datapoint {
	canID := int(binary.LittleEndian.Uint16(p.PacketBuffer[4:6]))
	config := canDatatypes[canID]
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
