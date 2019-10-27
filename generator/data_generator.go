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
	remote := false

	if len(os.Args) > 1 && os.Args[1] == "remote" {
		host = "solarracing.me"
		remote = true
	} else {
		host = "server"
	}

	log.Printf("Attempting to send data to host %s:6001\n", host)

	tcpConn, err := net.Dial("tcp", fmt.Sprintf("%s:6001", host))
	if err != nil {
		log.Fatalf("Error connecting to port: %s", err)
	}

	udpConn, err := net.Dial("udp", fmt.Sprintf("%s:6001", host))
	if err != nil {
		log.Fatalf("Error connecting to port: %s", err)
	}

	go listen(tcpConn)

	// every 10 seconds, send driver ack status (TCP/Reliable)
	go sendDriverStatuses(tcpConn)

	if !remote {
		go sendLocations(tcpConn)
	}

	// every 50 milliseconds, send test computation (UDP/Unreliable)
	sendTestComputation(udpConn)
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
		time.Sleep(10 * time.Second)
	}
}

type latlng struct {
	lat float32
	lng float32
}

func sendLocations(conn net.Conn) {
	// small loop on Hemphill
	points := []latlng{{33.786239, -84.406763}, {33.786235, -84.406200}, {33.785794, -84.406061}, {33.784849, -84.406088}, {33.785527, -84.406560}}
	i := 0
	for {
		packet := formPacket()
		sendLock.Lock()
		packet[2] = 0x21
		packet[3] = 0x06
		binary.LittleEndian.PutUint32(packet[4:8], math.Float32bits(points[i].lat))
		binary.LittleEndian.PutUint32(packet[8:12], math.Float32bits(points[i].lng))
		_, err := conn.Write(packet)
		sendLock.Unlock()
		if err != nil {
			log.Fatalf("Error writing to connection: %s", err)
		}
		packet = formPacket()
		sendLock.Lock()
		packet[2] = 0x2e
		packet[3] = 0x06
		packet[4] = 1
		_, err = conn.Write(packet)
		sendLock.Unlock()
		if err != nil {
			log.Fatalf("Error writing to connection: %s", err)
		}
		i = (i + 1) % len(points)
		time.Sleep(2 * time.Second)
	}
}

func formPacket() []byte {
	packet := append([]byte("GT"), make([]byte, 10)...)
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
