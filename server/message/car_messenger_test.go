package message

import (
	"reflect"
	"testing"
)

type FakeWriter struct {
	msg []byte
}

func (w *FakeWriter) Write(msg []byte) {
	w.msg = msg

}

func TestUploadChatMessageViaTCP(t *testing.T) {
	tcpPrefix := "GT"
	for _, tc := range []struct {
		title       string
		msg         string
		expectedMsg []byte
	}{{
		title:       "Basic message",
		msg:         "Go 40",
		expectedMsg: []byte{'G', 'T', 5, 'G', 'o', ' ', '4', '0'},
	}, {
		title:       "Empty message",
		msg:         "",
		expectedMsg: []byte{'G', 'T', 0},
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
