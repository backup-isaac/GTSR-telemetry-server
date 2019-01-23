package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

var sendLock sync.Mutex

func main() {
	var host string
	if len(os.Args) > 1 && os.Args[1] == "remote" {
		host = "solarracing.me"
	} else {
		host = "server"
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:6001", host))
	if err != nil {
		log.Fatalf("Error connecting to port: %s", err)
	}
	go listen(conn)
	// every 10 seconds, send driver ack status
	go sendDriverStatuses(conn)
	// every 50 milliseconds, send test computation
	sendTestComputation(conn)
}

func sendTestComputation(conn net.Conn) {
	for {
		packet := formPacket()
		sendLock.Lock()
		packet[2] = 0xFF
		packet[3] = 0xFF
		value := rand.Float32() * 100
		valueBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(valueBytes, math.Float32bits(value))
		for i := 0; i < 4; i++ {
			packet[4+i] = valueBytes[i]
		}
		_, err := conn.Write(packet)
		sendLock.Unlock()
		if err != nil {
			log.Fatalf("Error writing to connection: %s", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func sendDriverStatuses(conn net.Conn) {
	for {
		packet := formPacket()
		sendLock.Lock()
		packet[2] = 0x05
		packet[3] = 0x07
		packet[4] = 0x00
		_, err := conn.Write(packet)
		sendLock.Unlock()
		if err != nil {
			log.Fatalf("Error writing to connection: %s", err)
		}
		time.Sleep(10000 * time.Millisecond)
	}
}

func formPacket() []byte {
	packet := append([]byte("GTSR"), make([]byte, 12)...)
	return packet
}

// Receive dashboard messages.
func listen(conn net.Conn) {
	buf := make([]byte, 128)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("Error reading from connection: %s", err)
		}
		log.Printf("Received message from server: %q", buf[:n])
	}
}
