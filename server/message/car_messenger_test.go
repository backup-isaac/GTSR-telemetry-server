package message

import (
	"reflect"
	"testing"

	"server/datatypes"
)

const tcpPrefix = "GT"

type FakeWriter struct {
	msg []byte
}

func (w *FakeWriter) Write(msg []byte) {
	w.msg = msg

}

func TestUploadChatMessageViaTCP(t *testing.T) {
	for _, tc := range []struct {
		title       string
		msg         string
		expectedMsg []byte
	}{{
		title:       "Basic message",
		msg:         "Go 40",
		expectedMsg: []byte{'G', 'T', 'c', 5, 'G', 'o', ' ', '4', '0'},
	}, {
		title:       "Empty message",
		msg:         "",
		expectedMsg: []byte{'G', 'T', 'c', 0},
	}} {
		t.Run(tc.title, func(t *testing.T) {
			w := &FakeWriter{}
			m := NewCarMessenger(tcpPrefix, w)
			m.UploadChatMessageViaTCP(tc.msg)
			if !reflect.DeepEqual(tc.expectedMsg, w.msg) {
				t.Errorf("Unexpected message: want %q, got %q", tc.expectedMsg, w.msg)

			}

		})

	}

}

func TestUploadTCPPointMessage(t *testing.T) {
	// Building expected messages
	basicExpectedMsg := []byte{'G', 'T', 'd'}
	basicExpectedMsg = append(basicExpectedMsg, byte(42))
	basicExpectedMsg = append(basicExpectedMsg, convertFloat64to32(24.24)...)
	basicExpectedMsg = append(basicExpectedMsg, convertFloat64to32(25.25)...)
	basicExpectedMsg = append(basicExpectedMsg, convertFloat64to32(88)...)

	zerothPointExpectedMsg := []byte{'G', 'T', 'd'}
	zerothPointExpectedMsg = append(zerothPointExpectedMsg, byte(0))
	zerothPointExpectedMsg = append(zerothPointExpectedMsg, convertFloat64to32(24.24)...)
	zerothPointExpectedMsg = append(zerothPointExpectedMsg, convertFloat64to32(25.25)...)
	zerothPointExpectedMsg = append(zerothPointExpectedMsg, convertFloat64to32(88)...)

	zeroSpeedExpectedMsg := []byte{'G', 'T', 'd'}
	zeroSpeedExpectedMsg = append(zeroSpeedExpectedMsg, byte(42))
	zeroSpeedExpectedMsg = append(zeroSpeedExpectedMsg, convertFloat64to32(24.24)...)
	zeroSpeedExpectedMsg = append(zeroSpeedExpectedMsg, convertFloat64to32(25.25)...)
	zeroSpeedExpectedMsg = append(zeroSpeedExpectedMsg, convertFloat64to32(0)...)

	for _, tc := range []struct {
		title       string
		point       *datatypes.RoutePoint
		pointNumber int
		expectedMsg []byte
	}{{
		title: "Basic point",
		point: &datatypes.RoutePoint{
			Latitude:  24.24,
			Longitude: 25.25,
			Speed:     88,
		},
		pointNumber: 42,
		expectedMsg: basicExpectedMsg,
	}, {
		title: "0th point",
		point: &datatypes.RoutePoint{
			Latitude:  24.24,
			Longitude: 25.25,
			Speed:     88,
		},
		pointNumber: 0,
		expectedMsg: zerothPointExpectedMsg,
	}, {
		title: "Zero speed",
		point: &datatypes.RoutePoint{
			Latitude:  24.24,
			Longitude: 25.25,
			Speed:     0,
		},
		pointNumber: 42,
		expectedMsg: zeroSpeedExpectedMsg,
	}} {
		t.Run(tc.title, func(t *testing.T) {
			w := &FakeWriter{}
			m := NewCarMessenger(tcpPrefix, w)
			m.UploadTCPPointMessage(tc.point, tc.pointNumber)
			if !reflect.DeepEqual(tc.expectedMsg, w.msg) {
				t.Errorf("Unexpected message: want %q, got %q", tc.expectedMsg, w.msg)
			}
		})
	}
}
