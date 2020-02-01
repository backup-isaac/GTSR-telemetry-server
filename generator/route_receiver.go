package main

import (
	"encoding/binary"
	"fmt"
	"math"
)

// RoutePoint represents a suggested speed datapoint on the route
type RoutePoint struct {
	Latitude  float32
	Longitude float32
	Speed     float32
}

func (r *RoutePoint) String() string {
	return fmt.Sprintf("Latitude: %f, Longitude: %f, Speed: %f", r.Latitude, r.Longitude, r.Speed)
}

type receiverState byte

const (
	awaitingPreamble  receiverState = 0
	receivingNewRoute receiverState = 1
	receivingPoint    receiverState = 2
)

// RouteReceiver handles route transactions with the server
type RouteReceiver struct {
	Route    []*RoutePoint
	preamble []byte
	buf      []byte
	state    receiverState
}

// NewRouteReceiver creates a new RouteReceiver
func NewRouteReceiver() *RouteReceiver {
	return &RouteReceiver{
		preamble: make([]byte, 3),
	}
}

// ReceiveByte processes a given byte from the data stream. If
// a full packet is recognized
func (r *RouteReceiver) ReceiveByte(b byte) ([]byte, bool) {
	switch r.state {
	case awaitingPreamble:
		r.preamble = append(r.preamble[1:], b)
		if string(r.preamble) == "GTt" {
			r.state = receivingNewRoute
		} else if string(r.preamble) == "GTd" {
			r.state = receivingPoint
			r.buf = []byte{}
		}
	case receivingNewRoute:
		r.Route = make([]*RoutePoint, b)
		r.state = awaitingPreamble
		return trackAck(), true
	case receivingPoint:
		r.buf = append(r.buf, b)
		if len(r.buf) >= 13 {
			var resp []byte
			var ok bool
			if r.buf[0] >= 0 && int(r.buf[0]) < len(r.Route) {
				r.Route[r.buf[0]] = &RoutePoint{
					Latitude:  parseFloat32(r.buf[1:5]),
					Longitude: parseFloat32(r.buf[5:9]),
					Speed:     parseFloat32(r.buf[9:13]),
				}
				resp = packetAck(r.buf[0])
				ok = true
			}
			r.state = awaitingPreamble
			return resp, ok
		}
	}
	return nil, false
}

// RouteReceived returns whether a complete route has been received
func (r *RouteReceiver) RouteReceived() bool {
	if len(r.Route) == 0 {
		return false
	}
	for _, point := range r.Route {
		if point == nil {
			return false
		}
	}
	return true
}

func parseFloat32(buf []byte) float32 {
	rawValue := binary.LittleEndian.Uint32(buf)
	return math.Float32frombits(rawValue)
}

func trackAck() []byte {
	return append([]byte{
		'G', 'T',
		0x0A, 0x07,
	}, make([]byte, 8)...)
}

func packetAck(packetNum byte) []byte {
	return append([]byte{
		'G', 'T',
		0x0B, 0x07,
		packetNum,
	}, make([]byte, 7)...)
}
