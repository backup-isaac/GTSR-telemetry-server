package listener

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.gatech.edu/GTSR/telemetry-server/canConfigs"
	"github.gatech.edu/GTSR/telemetry-server/datatypes"
)

const (
	connHost = "localhost"
	connPort = "6001"
	connType = "tcp"
)

// ConnectionHandler is the object representing the TCP listener
type ConnectionHandler struct {
	Publisher DatapointPublisher
	Parser    PacketParser
}

// NewConnectionHandler returns an initialized ConnectionHandler
func NewConnectionHandler(publisher DatapointPublisher, parser PacketParser) *ConnectionHandler {
	return &ConnectionHandler{
		Publisher: publisher,
		Parser:    parser,
	}
}

var connections sync.Map

// HandleConnection handles a new connection
func (handler *ConnectionHandler) HandleConnection(conn net.Conn) {
	connections.Store(conn.RemoteAddr().String(), conn)
	defer connections.Delete(conn.RemoteAddr().String())
	buf := make([]byte, 1024)
	for {
		reqLen, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Fatalf("Error reading from %s: %s", conn.RemoteAddr().Network(), err)
		}
		for i := 0; i < reqLen; i++ {
			if handler.Parser.ParseByte(buf[i]) {
				points := handler.Parser.ParsePacket()
				for _, point := range points {
					handler.Publisher.Publish(point)
				}
			}
		}
	}
}

// Listen is the main function of listener which listens to the TCP data port for incoming connections
func Listen() {
	canConfigs, err := canConfigs.LoadConfigs()
	if err != nil {
		log.Fatalf("Error loading CAN configs: %s", err)
	}
	connListener, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		log.Fatalf("Error listening on TCP port: %s", err)
	}
	defer connListener.Close()
	fmt.Printf("Listening on %s:%s\n", connHost, connPort)
	consecutiveFailures := 0
	for {
		conn, err := connListener.Accept()
		if err == nil {
			consecutiveFailures = 0
			fmt.Println("Received connection from", conn.RemoteAddr().String())
			handler := NewConnectionHandler(NewDatapointPublisher(), NewPacketParser(canConfigs))
			go handler.HandleConnection(conn)
		} else {
			consecutiveFailures++
			fmt.Println("Error accepting connection in function Listen: listener/listener.go")
			fmt.Printf("Consecutive connection failures: %d\n", consecutiveFailures)
			if consecutiveFailures >= 5 {
				log.Fatal("Consecutive connection failures exceeded maximum limit")
			}
		}
	}
}

// Write writes the data in buf to all open connections
func Write(buf []byte) {
	connections.Range(func(key, value interface{}) bool {
		conn := value.(net.Conn)
		conn.Write(buf)
		return true
	})
}

// Subscribe subscribes the channel c to the datapoint publisher
func Subscribe(c chan *datatypes.Datapoint) error {
	return NewDatapointPublisher().Subscribe(c)
}

// Unsubscribe unsubscribes the channel c from the datapoint publisher
func Unsubscribe(c chan *datatypes.Datapoint) error {
	return NewDatapointPublisher().Unsubscribe(c)
}
