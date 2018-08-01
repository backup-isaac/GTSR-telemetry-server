package listener

import (
	"fmt"
	"log"
	"net"
)

const (
	connHost   = "localhost"
	connPort   = "6001"
	connType   = "tcp"
	dataLength = 16
)

// Listener is the object representing the TCP listener
type Listener struct {
	publisher DatapointPublisher
	parser    PacketParser
}

// NewListener returns an initialized Listener
func NewListener() *Listener {
	return &Listener{
		publisher: NewDatapointPublisher(),
		parser:    NewPacketParser(),
	}
}

// Datapoint is a container for raw data from the car
type Datapoint struct {
	// Metric is the name of the metric type for this datapoint
	// Examples: Wavesculptor RPM, BMS Current
	Metric string
	// Value of this datapoint
	Value interface{}
	// Map of tags associated with this datapoint (e.g. event tags)
	Tags map[string]string
}

// Listen is the main function of listener which listens to the TCP data port for incoming connections
func Listen() {
	connListener, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		log.Fatal(fmt.Errorf("Error listening on TCP port: %s", err))
	}
	defer connListener.Close()
	fmt.Printf("Listening on %s:%s\n", connHost, connPort)
	consecutiveFailures := 0
	for {
		conn, err := connListener.Accept()
		if err == nil {
			consecutiveFailures = 0
			listener := NewListener()
			go listener.HandleRequest(conn)
		} else {
			consecutiveFailures++
			fmt.Println("Error accepting connection in function Listen: listener/listen.go")
			fmt.Printf("Consecutive connection failures: %d\n", consecutiveFailures)
			if consecutiveFailures >= 5 {
				log.Fatal("Consecutive connection failures exceeded maximum limit")
			}
		}
	}
}

// HandleRequest handles a new connection
func (listener *Listener) HandleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Error reading from %s: %s", conn.RemoteAddr().Network(), err)
	}
	for i := 0; i < reqLen; i++ {
		if listener.parser.ParseByte(buf[i]) {
			point := listener.parser.ParsePacket()
			listener.publisher.Publish(point)
		}
	}
}
