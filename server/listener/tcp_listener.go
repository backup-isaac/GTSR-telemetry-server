package listener

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"server/storage"
	"sync"
	"sync/atomic"
	"time"

	"server/configs"
	"server/datatypes"
)

const (
	connHost = "0.0.0.0"
	connPort = 6001
	connType = "tcp"
)

// TCPConnectionHandler is the object representing the TCP listener
type TCPConnectionHandler struct {
	Publisher DatapointPublisher
	Parser    PacketParser
}

// NewTCPConnectionHandler returns an initialized TCPConnectionHandler
func NewTCPConnectionHandler(publisher DatapointPublisher, parser PacketParser) *TCPConnectionHandler {
	return &TCPConnectionHandler{
		Publisher: publisher,
		Parser:    parser,
	}
}

var connections sync.Map
var activeConnectionCount uint32

func reportConnections() {
	store, err := storage.NewStorage()
	if err != nil {
		log.Println("Error getting storage for connection reporting.")
		return
	}
	defer store.Close()
	ticker := time.NewTicker(time.Second * 5)
	for {
		<-ticker.C
		store.Insert([]*datatypes.Datapoint{
			&datatypes.Datapoint{
				Metric: "Active_TCP_Connections",
				Value:  float64(atomic.LoadUint32(&activeConnectionCount)),
				Time:   time.Now(),
			},
		})
	}
}

// HandleTCPConnection handles a new connection
func (handler *TCPConnectionHandler) HandleTCPConnection(conn net.Conn) {
	defer conn.Close()
	connectionKey := conn.RemoteAddr().String() + ";" + string(rand.Intn(1000000))
	connections.Store(connectionKey, conn)
	atomic.AddUint32(&activeConnectionCount, 1)
	defer connections.Delete(connectionKey)
	defer atomic.AddUint32(&activeConnectionCount, ^uint32(0)) // This is the documented way to decrement a uint atomically
	buf := make([]byte, 1024)
	for {
		reqLen, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from %s: %s\n", conn.RemoteAddr().String(), err)
			}
			log.Printf("Connection to %s lost\n", conn.RemoteAddr().String())
			return
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

// TCPListen is the main function of listener which listens to the TCP data port for incoming connections
func TCPListen() {
	canConfigs, err := configs.LoadConfigs()
	if err != nil {
		log.Fatalf("Error loading CAN configs: %s", err)
	}
	connListener, err := net.Listen(connType, fmt.Sprintf("%s:%d", connHost, connPort))
	if err != nil {
		log.Fatalf("Error listening on TCP port: %s", err)
	}
	defer connListener.Close()
	log.Printf("Listening on %s:%d TCP\n", connHost, connPort)
	consecutiveFailures := 0
	for {
		conn, err := connListener.Accept()
		if err == nil {
			consecutiveFailures = 0
			log.Println("Received connection from", conn.RemoteAddr().String())
			handler := NewTCPConnectionHandler(GetDatapointPublisher(), NewPacketParser(canConfigs))
			go handler.HandleTCPConnection(conn)
		} else {
			consecutiveFailures++
			log.Println("Error accepting connection in function Listen: listener/listener.go")
			log.Printf("Consecutive connection failures: %d\n", consecutiveFailures)
			if consecutiveFailures >= 5 {
				log.Fatal("Consecutive connection failures exceeded maximum limit")
			}
		}
	}
}

var writeChannel = make(chan []byte, 100)

// Write writes the data in buf to all open connections
func Write(buf []byte) {
	writeChannel <- append(make([]byte, 0, len(buf)), buf...)
}

func writerThread() {
	for {
		buf := <-writeChannel
		connections.Range(func(key, value interface{}) bool {
			conn := value.(net.Conn)
			_, err := conn.Write(buf)
			if err != nil {
				conn.Close()
				log.Printf("Error writing to %s - closing\n", conn.RemoteAddr().String())
			}
			return true
		})
	}
}

// Subscribe subscribes the channel c to the datapoint publisher
func Subscribe(c chan *datatypes.Datapoint) error {
	return GetDatapointPublisher().Subscribe(c)
}

// Unsubscribe unsubscribes the channel c from the datapoint publisher
func Unsubscribe(c chan *datatypes.Datapoint) error {
	return GetDatapointPublisher().Unsubscribe(c)
}

func init() {
	go reportConnections()
	go writerThread()
}
