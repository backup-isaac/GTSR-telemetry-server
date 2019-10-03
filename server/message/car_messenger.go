package message

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"sync"

	"server/datatypes"
)

var routePointsMutex = sync.Mutex{}

const routePointsJSONPath = "../map/route.json"

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
	constructedMsg = append(constructedMsg, byte(len(message)))
	constructedMsg = append(constructedMsg, []byte(message)...)

	m.Writer.Write(constructedMsg)
}

// UploadTrackInfoViaTCP sends the server's copy of track info to the car
func (m *CarMessenger) UploadTrackInfoViaTCP() error {
	routePointsMutex.Lock()
	defer routePointsMutex.Unlock()

	routePointsFile, err := ioutil.ReadFile(routePointsJSONPath)
	if err != nil {
		return errors.New("Error reading map/route.json to send new points to dashboard: " + err.Error())
	}

	var points []datatypes.RoutePoint
	json.Unmarshal(routePointsFile, &points)

	// Send a message telling the dashboard how many new points will be sent
	constructedMsg := make([]byte, 0)

	constructedMsg = append(constructedMsg, []byte(m.TCPPrefix)...)
	constructedMsg = append(constructedMsg, byte(NumIncomingDatapoints))
	constructedMsg = append(constructedMsg, byte(len(points)))

	m.Writer.Write(constructedMsg)

	// Send each point from map/route.json to the dashboard
	for i, point := range points {
		m.UploadTCPPointMessage(&point, i)

		// Listen for ACK from the car

		// If we don't get an ACK within a timeout, subtract 1 from i so that
		// when the loop increments, we attempt to send the same point again
	}

	return nil
}

// UploadTCPPointMessage sends a new track info point to the dashboard
// Message protocol currently looks like: GTSR_d_#_distance_latitude_longitude_speed
func (m *CarMessenger) UploadTCPPointMessage(p *datatypes.RoutePoint, pointNumber int) {
	constructedMsg := make([]byte, 0)

	constructedMsg = append(constructedMsg, []byte(m.TCPPrefix)...)
	constructedMsg = append(constructedMsg, byte(DatapointClassifier))
	constructedMsg = append(constructedMsg, byte(pointNumber))
	constructedMsg = append(constructedMsg, convertFloat64to32(p.Distance)...)
	constructedMsg = append(constructedMsg, convertFloat64to32(p.Latitude)...)
	constructedMsg = append(constructedMsg, convertFloat64to32(p.Longitude)...)
	constructedMsg = append(constructedMsg, convertFloat64to32(p.Speed)...)

	m.Writer.Write(constructedMsg)
}

func convertFloat64to32(num float64) []byte {
	num32 := float32(num)
	bits := math.Float32bits(num32)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, bits)
	return buf
}
