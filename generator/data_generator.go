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

	// simulate Wavesculptor telemetry as if the car is driving
	go driveMotors(tcpConn)

	// every 50 milliseconds, send test computation (UDP/Unreliable)
	sendTestComputation(udpConn)
}

// pretend that the car drives a couple seconds, then stops for a bit, then repeats
func driveMotors(conn net.Conn) {
	leftMotorRpms := []float32{
		0.0, 3.8, 8.5, 13.0, 18.0, 23.0, 28.4, 34.1, 40.0, 45.0, 49.2, 44.8, 44.1, 37.6, 31.8, 25.3, 19.9, 14.7, 10.1, 6.2, 2.7, 0.0, 0.0,
	}
	rightMotorRpms := []float32{
		0.0, 3.8, 8.4, 13.0, 18.2, 23.1, 28.4, 34.0, 39.7, 44.9, 49.3, 44.8, 44.4, 37.7, 31.7, 25.3, 19.7, 14.5, 10.1, 6.3, 2.7, 0.1, 0.0,
	}
	i := 0
	for {
		var leftMotorRpm, rightMotorRpm float32
		if i < len(leftMotorRpms) {
			leftMotorRpm = leftMotorRpms[i]
			rightMotorRpm = rightMotorRpms[i]
		} else {
			leftMotorRpm = 0
			rightMotorRpm = 0
		}
		err := sendFloatPacket(0x423, leftMotorRpm, conn)
		if err != nil {
			log.Fatalf("Error writing to connection: %s", err)
		}
		err = sendFloatPacket(0x403, rightMotorRpm, conn)
		if err != nil {
			log.Fatalf("Error writing to connection: %s", err)
		}
		i = (i + 1) % 40
		time.Sleep(250 * time.Millisecond)
	}
}

func sendFloatPacket(id uint16, value float32, conn net.Conn) error {
	packet := formPacket()
	packet[2] = byte(id & 0xff)
	packet[3] = byte((id & 0xff00) >> 8)
	binary.LittleEndian.PutUint32(packet[4:8], math.Float32bits(value))
	sendLock.Lock()
	_, err := conn.Write(packet)
	sendLock.Unlock()
	return err
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
	receiver := NewRouteReceiver()
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("Error reading from connection: %s", err)
		}
		log.Printf("Received message from server: %q", buf[:n])
		for i := 0; i < n; i++ {
			if response, ok := receiver.ReceiveByte(buf[i]); ok && rand.Float32() < 0.7 {
				// Full packet received. Randomly drop some responses to test robustness
				sendLock.Lock()
				_, err := conn.Write(response)
				sendLock.Unlock()
				if err != nil {
					log.Fatalf("Error writing to connection: %+v", err)
				}
				if receiver.RouteReceived() {
					log.Println("Full route received:")
					for _, point := range receiver.Route {
						log.Println("\t" + point.String())
					}
				}
			}
		}
	}
}
