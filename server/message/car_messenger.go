package message

import (
	"encoding/binary"
	"math"
	"server/datatypes"
)

// Writer handles writing a message to a TCP listener
type Writer interface {
	Write([]byte)
}

// CarMessenger handles sending messages to the car
type CarMessenger struct {
	TCPPrefix string
	Writer    Writer
}

// NewCarMessenger returns a new Messenger initialized with the provided TCP
// prefix that will write new messages to the provided Writer
func NewCarMessenger(tcpPrefix string, writer Writer) *CarMessenger {
	return &CarMessenger{
		TCPPrefix: tcpPrefix,
		Writer:    writer,
	}
}

// UploadChatMessageViaTCP sends the provided message to the listener, which
// will then relay it to the car
func (m *CarMessenger) UploadChatMessageViaTCP(message string) {
	constructedMsg := make([]byte, 0)

	constructedMsg = append(constructedMsg, []byte(m.TCPPrefix)...)
	constructedMsg = append(constructedMsg, byte(slackMessage))
	constructedMsg = append(constructedMsg, byte(len(message)))
	constructedMsg = append(constructedMsg, []byte(message)...)

	m.Writer.Write(constructedMsg)
}

// UploadNewRoute sends a New Route message to the car
func (m *CarMessenger) UploadNewRoute(len int) {
	m.Writer.Write(append([]byte(m.TCPPrefix), byte(routeBegin), byte(len)))
}

// UploadTCPPointMessage sends a new track info point to the dashboard
// Message protocol currently looks like: GTSR_d_#_latitude_longitude_speed
func (m *CarMessenger) UploadTCPPointMessage(p *datatypes.RoutePoint, pointNumber int) {
	constructedMsg := make([]byte, 0)

	constructedMsg = append(constructedMsg, []byte(m.TCPPrefix)...)
	constructedMsg = append(constructedMsg, byte(dataPoint))
	constructedMsg = append(constructedMsg, byte(pointNumber))
	constructedMsg = append(constructedMsg, convertFloat64to32(p.Latitude)...)
	constructedMsg = append(constructedMsg, convertFloat64to32(p.Longitude)...)
	constructedMsg = append(constructedMsg, convertFloat64to32(p.Speed)...)

	m.Writer.Write(constructedMsg)
}

func convertFloat64to32(num float64) []byte { //probably want to convert to fewer bits
	num32 := float32(num)
	bits := math.Float32bits(num32)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, bits)
	return buf
}
