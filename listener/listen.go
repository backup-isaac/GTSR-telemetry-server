package listener

import (
	"fmt"
	"log"
	"net"
	"sync"
)

const (
	connHost   = "localhost"
	connPort   = "6001"
	connType   = "tcp"
	dataLength = 16
)

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

func init() {
	initPublisher()
	err := loadConfigs()
	if err != nil {
		log.Fatal(err)
	}
}

func initPublisher() {
	if publisher != nil {
		return
	}
	publisher = &DatapointPublisher{
		Subscribers:     []chan *Datapoint{},
		SubscribersLock: &sync.Mutex{},
		PublishChannel:  make(chan *Datapoint),
	}
	go publisherThread()
}

// Listen is the main function of listener which listens to the TCP data port for incoming connections
func Listen() {
	listener, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		log.Fatal(fmt.Errorf("Error listening on TCP port: %s", err))
	}
	defer listener.Close()
	fmt.Printf("Listening on %s:%s\n", connHost, connPort)
	consecutiveFailures := 0
	for {
		conn, err := listener.Accept()
		if err == nil {
			consecutiveFailures = 0
			go handleRequest(conn)
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

func handleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error reading from %s: %s", conn.RemoteAddr().Network(), err))
	}
	fmt.Println(string(buf[:reqLen]))
}
