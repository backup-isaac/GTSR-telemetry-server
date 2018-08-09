package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:6001")
	if err != nil {
		log.Fatalf("Error connecting to port: %s", err)
	}
	go listen(conn)
	packet := formPacket()
	for {
		value := rand.Float32() * 100
		valueBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(valueBytes, math.Float32bits(value))
		for i := 0; i < 4; i++ {
			packet[8+i] = valueBytes[i]
		}
		_, err := conn.Write(packet)
		if err != nil {
			log.Fatalf("Error writing to connection: %s", err)
		}
		<-time.After(50 * time.Millisecond)
	}
}

func formPacket() []byte {
	packet := append([]byte("GTSR"), make([]byte, 12)...)
	packet[4] = 0xFF
	packet[5] = 0xFF
	return packet
}

func listen(conn net.Conn) {
	isWriting := false
	buf := make([]byte, 4)
	valOffset := 0
	for {
		conn.Read(buf)
		if string(buf) == "GTSR" {
			isWriting = !isWriting
			valOffset = 0
		} else {
			if !isWriting {
				continue
			}
			bits := binary.LittleEndian.Uint32(buf)
			val := math.Float32frombits(bits)
			switch valOffset {
			case 0:
				fmt.Printf("Distance: %v\n", val)
			case 1:
				fmt.Printf("Latitude: %v\n", val)
			case 2:
				fmt.Printf("Longitude: %v\n", val)
			case 3:
				fmt.Printf("Speed: %v\n\n", val)
			}
			valOffset = (valOffset + 1) % 4
		}
	}
}
