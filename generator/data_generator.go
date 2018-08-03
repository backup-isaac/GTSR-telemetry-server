package main

import (
	"encoding/binary"
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
