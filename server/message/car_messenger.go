package message

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"sync"
	"time"

	"server/datatypes"
	"server/listener"

	"github.com/nlopes/slack"
)

var routePointsMutex = sync.Mutex{}

const routePointsJSONPath = "../map/route.json"

// # of times we attempt to send a message to the car if we don't receive ACKs
const retryAttempts = 5

// Length of time in seconds we wait for the car to respond with an ACK
const timeoutLen = 3

var slck *slack.Client
var slackMessenger = NewSlackMessenger(slck)

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
	constructedMsg = append(constructedMsg, byte(SlackMessageClassifier))
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

	// Subscribe to datapoint publisher to listen for ACKs in response to
	// messages that we sent below
	pointsFromCar := make(chan *datatypes.Datapoint)
	listener.Subscribe(pointsFromCar)

	// Send a message telling the dashboard how many new points will be sent
	curSendAttempts := 0
	didTelemBoardAck := false
	for !didTelemBoardAck && curSendAttempts < retryAttempts {
		// Construct & send the "heads-up" message
		constructedMsg := make([]byte, 0)
		constructedMsg = append(constructedMsg, []byte(m.TCPPrefix)...)
		constructedMsg = append(constructedMsg, byte(NumIncomingDataPointsClassifier))
		constructedMsg = append(constructedMsg, byte(len(points)))
		m.Writer.Write(constructedMsg)

		curSendAttempts++

		// Determine if we need to retry sending this "heads-up" message
		timer := time.NewTimer(timeoutLen * time.Second)
		<-timer.C
		select {
		case p := <-pointsFromCar:
			if p.Metric == "Receive_New_Track_Info_ACK_Status" && p.Value == 1.0 {
				timer.Stop()
				didTelemBoardAck = true
				slackMessenger.PostNewMessage("Server: ACK'd new track info \"heads-up\" message")
			}
		case <-timer.C:
			msg := "New track info \"heads-up\" message: timeout. "
			if curSendAttempts < retryAttempts {
				msg += fmt.Sprintf("Retrying... (%q of %q)", curSendAttempts, retryAttempts)
			}
		}
	}

	if !didTelemBoardAck {
		return errors.New("Reached max number of retry attempts while sending new track info \"heads-up\" message")
		slackMessenger.PostNewMessage("Reached max number of retry attempts.")
	}

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
// Message protocol currently looks like: GTSR_d_#_latitude_longitude_speed
func (m *CarMessenger) UploadTCPPointMessage(p *datatypes.RoutePoint, pointNumber int) {
	constructedMsg := make([]byte, 0)

	constructedMsg = append(constructedMsg, []byte(m.TCPPrefix)...)
	constructedMsg = append(constructedMsg, byte(DataPointClassifier))
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
