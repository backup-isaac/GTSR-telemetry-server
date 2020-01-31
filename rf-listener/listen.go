package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"

	"hash/crc32"

	"go.bug.st/serial.v1"
)

func main() {
	var host string
	var serialPort string
	var table = crc32.MakeTable(0x1EDC6F41)

	// connect to the serial port
	if len(os.Args) > 1 {
		serialPort = os.Args[1]
	} else {
		// TODO: Mock Serial Connection?
		ports, err := serial.GetPortsList()
		if err != nil {
			log.Fatal(err)
		}
		if len(ports) == 0 {
			log.Fatal("No serial ports found!")
		}
		for _, port := range ports {
			log.Printf("Found port: %v\n", port)
		}
		log.Fatalf("Please specify serial port (e.g. COM4 or /dev/ttyUSB0)")
	}

	c := &serial.Mode{BaudRate: 115200}
	s, err := serial.Open(serialPort, c)
	if err != nil {
		log.Fatalf("Serial Error: %s", err)
	} else {
		log.Printf("Successfully connected to %s\n", serialPort)
	}
	defer s.Close()

	// set if we are uploading to the production server or localhost
	if len(os.Args) > 2 && os.Args[2] == "remote" {
		host = "solarracing.me"
	} else if len(os.Args) > 2 {
		host = os.Args[2]
	} else {
		log.Println("Argument \"remote\" not specified. Relaying to localhost.")
		host = "localhost"
	}

	log.Printf("Attempting to connect to %s\n", host)

	// attempt to listen to the server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:6001", host))
	if err != nil {
		log.Fatalf("Error connecting to port: %s", err)
	} else {
		log.Printf("Successfully connected to %s\n", host)
	}
	defer conn.Close()

	// listen for incoming TCP messages, and print out
	go listen(conn, s)

	// receive messages from serial port
	buf := make([]byte, 144)
	for {
		n, err := s.Read(buf)
		if err != nil {
			log.Fatalf("Serial error: %s", err)
		}
		// use CRC to verify message, if specified
		if len(os.Args) > 3 && os.Args[3] == "CRC" {
			// create a new buffer to add uncompromised frames to
			cleanBuf := make([]byte, 144)
			j := 0
			// size of a frame from TelemBoard is 12 bytes + 4 byte CRC checksum
			for i := 0; i < n; i += 16 {
				// if CRC fails, do nothing
				if !verifyChecksum(buf[i:i+16], table) {
					log.Println("CRC failed.")
					continue
				}
				log.Println("CRC passed!")
				copy(cleanBuf[j:j+12], buf[i:i+12])
				j += 12
			}
			// replace buf and n with values after CRC
			buf = cleanBuf
			n = j
		}
		// directly relay messages from serial to tcp
		_, err = conn.Write(buf[:n])
		if err != nil {
			log.Fatalf("Error writing to connection: %s", err)
		}
	}
}

// Dashboard messages and prints them out
func listen(conn net.Conn, s serial.Port) {
	buf := make([]byte, 128)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("Error reading from connection: %s", err)
		}
		log.Printf("Received message from server: %q", buf[:n])

		// relay the message via serial
		_, err = s.Write(buf[:n])
		if err != nil {
			log.Fatalf("Error writing to Serial Port :%s", err)
		}
	}
}

func verifyChecksum(buf []byte, table *crc32.Table) bool {
	checksumTransmitted := binary.LittleEndian.Uint32(buf[len(buf)-4:])
	checksumCalculated := crc32.Checksum(buf[:len(buf)-4], table)
	return checksumTransmitted == checksumCalculated
}
