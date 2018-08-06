package listener

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.gatech.edu/GTSR/telemetry-server/canConfigs"
	"github.gatech.edu/GTSR/telemetry-server/datatypes"
)

const (
	connHost = "localhost"
	connPort = "6001"
	connType = "tcp"
)

// Listener is the object representing the TCP listener
type Listener struct {
	Publisher DatapointPublisher
	Parser    PacketParser
}

// NewListener returns an initialized Listener
func NewListener(publisher DatapointPublisher, parser PacketParser) *Listener {
	return &Listener{
		Publisher: publisher,
		Parser:    parser,
	}
}

// HandleRequest handles a new connection
func (listener *Listener) HandleRequest(conn net.Conn) {
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
			if listener.Parser.ParseByte(buf[i]) {
				points := listener.Parser.ParsePacket()
				for _, point := range points {
					listener.Publisher.Publish(point)
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
			listener := NewListener(NewDatapointPublisher(), NewPacketParser(canConfigs))
			go listener.HandleRequest(conn)
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

// Subscribe subscribes the channel c to the datapoint publisher
func Subscribe(c chan *datatypes.Datapoint) error {
	return NewDatapointPublisher().Subscribe(c)
}

// Unsubscribe unsubscribes the channel c from the datapoint publisher
func Unsubscribe(c chan *datatypes.Datapoint) error {
	return NewDatapointPublisher().Unsubscribe(c)
}
