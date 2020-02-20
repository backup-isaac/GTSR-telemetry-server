package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"hash/crc32"

	"go.bug.st/serial.v1"
)

func main() {
	var host string
	var serialPort string

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
	go readWriteBytes(conn, s)

	// if CRC enabled, use a different algorithm to read in bytes
	if len(os.Args) > 3 && os.Args[3] == "CRC" {
		err = readWriteBytesCRC(s, conn)
		log.Fatalf("CRC-enabled read/write failed: %s", err)
	}
	// receive messages from serial port
	err = readWriteBytes(s, conn)
	log.Fatalf("Read/write failed: %s", err)
}

// Read from some io.Reader (e.g. the serial port or a file reader) and write the bytes to some io.Writer (e.g. the open socket or a file writer).
func readWriteBytes(reader io.Reader, writer io.Writer) error {
	buf := make([]byte, 128)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			return err
		}
		_, err = writer.Write(buf[:n])
		if err != nil {
			return err
		}
	}
}

func readWriteBytesCRC(reader io.Reader, writer io.Writer) error {
	buf := make([]byte, 144)
	packetIndex := 0
	packetBuffer := make([]byte, 16)
	table := crc32.MakeTable(0x1EDC6F41)
	for {
		numBytes, err := reader.Read(buf)
		if err != nil {
			return err
		}
		for i := 0; i < numBytes; i++ {
			parseByte(buf[i], &packetIndex, packetBuffer)
			if err != nil {
				continue
			}
			if packetIndex == len(packetBuffer) {
				packetIndex = 0
				if verifyChecksum(packetBuffer, table) {
					_, err := writer.Write(packetBuffer[:12])
					if err != nil {
						return err
					}
				}
			}
		}
	}
}

func verifyChecksum(buf []byte, table *crc32.Table) bool {
	checksumTransmitted := binary.LittleEndian.Uint32(buf[len(buf)-4:])
	checksumCalculated := crc32.Checksum(buf[:len(buf)-4], table)
	return checksumTransmitted == checksumCalculated
}

func parseByte(b byte, packetIndex *int, packetBuffer []byte) error {
	switch *packetIndex {
	case 0:
		if b != 'G' {
			return fmt.Errorf("G not found")
		}
		packetBuffer[*packetIndex] = b
		*packetIndex++
	case 1:
		if b != 'T' {
			*packetIndex = 0
			return fmt.Errorf("T not found")
		}
		packetBuffer[*packetIndex] = b
		*packetIndex++
	default:
		packetBuffer[*packetIndex] = b
		*packetIndex++
	}
	return nil
}
