package listener

import (
	"encoding/binary"
	"net"
	"server/configs"
	"server/datatypes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTCPConnectionHandler(t *testing.T) {
	parser := NewPacketParser(map[int][]*configs.CanConfigType{
		0x100: {{
			CanID:    0x100,
			Datatype: "int32",
			Name:     "Test1",
			Offset:   0,
		}, {
			CanID:    0x100,
			Datatype: "int32",
			Name:     "Test2",
			Offset:   4,
		}},
	})
	publisher := newDatapointPublisher()
	defer publisher.Close()
	c := make(chan *datatypes.Datapoint, 2)
	err := publisher.Subscribe(c)
	assert.NoError(t, err)
	l := NewTCPConnectionHandler(publisher, parser)
	server, client := net.Pipe()
	go l.HandleTCPConnection(client)
	server.Write([]byte{'G', 'T'})
	binary.Write(server, binary.LittleEndian, uint16(0x100))
	binary.Write(server, binary.LittleEndian, int32(12345))
	binary.Write(server, binary.LittleEndian, int32(54321))
	time.Sleep(100 * time.Millisecond)
	err = server.Close()
	assert.NoError(t, err)

	gotTest1 := false
	gotTest2 := false

	for i := 0; i < 2; i++ {
		select {
		case p := <-c:
			switch p.Metric {
			case "Test1":
				assert.Equal(t, float64(12345), p.Value)
				gotTest1 = true
			case "Test2":
				assert.Equal(t, float64(54321), p.Value)
				gotTest2 = true
			default:
				t.Fail()
			}
		default:
			t.Fail()
		}
	}

	assert.True(t, gotTest1)
	assert.True(t, gotTest2)
}
