package listener_test

import (
	"net"
	"testing"
	"time"

	"telemetry-server/datatypes"
	"telemetry-server/listener"
	"telemetry-server/listener/mocks"

	"github.com/stretchr/testify/assert"
)

func TestConnectionHandler(t *testing.T) {
	parser := &mocks.PacketParser{}
	parser.On("ParseByte", uint8(0)).Return(false)
	parser.On("ParseByte", uint8(1)).Return(true)
	datapoint := []*datatypes.Datapoint{{}}
	parser.On("ParsePacket").Return(datapoint)

	publisher := &mocks.DatapointPublisher{}
	publisher.On("Publish", datapoint[0]).Return()

	l := &listener.ConnectionHandler{
		Publisher: publisher,
		Parser:    parser,
	}
	server, client := net.Pipe()
	go l.HandleConnection(client)

	server.Write(make([]byte, 4))
	server.Write(make([]byte, 2))
	server.Write([]byte{0, 0, 0, 1, 0, 0})
	time.Sleep(100 * time.Millisecond)
	err := server.Close()
	assert.NoError(t, err)

	parser.AssertNumberOfCalls(t, "ParseByte", 12)
	parser.AssertNumberOfCalls(t, "ParsePacket", 1)
	publisher.AssertCalled(t, "Publish", datapoint[0])
}
