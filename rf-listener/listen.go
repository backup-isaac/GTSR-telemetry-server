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

	buf := make([]byte, 144)
	// if CRC enabled, use a different algorithm to read in bytes
	if len(os.Args) > 3 && os.Args[3] == "CRC" {
		bytesRead := 0
		incompleteBuf := make([]byte, 16)
		checkGT := true
		for {
			n, err := s.Read(buf)
			if err != nil {
				log.Fatalf("Serial error: %s", err)
			}
			// if this is the first iteration or CRC was failed in the previous iteration, look for a GT
			if checkGT {
				// if the first two characters between the incomplete buffer and read buffer aren't GT, reset bytesRead and look for a new GT in buf
				if (bytesRead == 0) || (bytesRead == 1 && (incompleteBuf[0] != 'G' || buf[0] != 'T')) || incompleteBuf[0] != 'G' || incompleteBuf[1] != 'T' {
					bytesRead = 0
					for i := 0; i < n; i++ {
						if buf[i] == 'G' {
							if i+1 >= n {
								incompleteBuf[0] = 'G'
								bytesRead = 1
							} else if buf[i+1] == 'T' {
								checkGT = false
								break
							}
						}
					}
					if checkGT {
						continue
					}
				} else {
					checkGT = false
				}
			}
			// check if we need to add bytes to an incomplete frame
			start := 0
			if bytesRead > 0 {
				if n+bytesRead < 16 {
					// add bytes to existing incomplete frame
					copy(incompleteBuf[bytesRead:bytesRead+n], buf[:])
					bytesRead += n
					continue
				} else {
					// finish incomplete frame, and offset start position in buf
					copy(incompleteBuf[bytesRead:16], buf[:16-bytesRead])
					start += 16 - bytesRead
					// reset bytesRead to 0
					bytesRead = 0
				}
			}
			// create a new buffer to add uncompromised frames to
			cleanBuf := make([]byte, 144)
			var i int
			j := 0
			// if incompleteBuf was finished, do CRC then add it to cleanBuf to be written if passed
			if start > 0 {
				if !verifyChecksum(incompleteBuf, table) {
					log.Println("CRC failed.")
					checkGT = true
				} else {
					log.Println("CRC passed!")
					copy(cleanBuf[:12], incompleteBuf[:12])
					j = 12
				}
			}
			// for each complete frame, do CRC, then add to cleanBuf if passed
			for i = start; i < n-15; i += 16 {
				if !verifyChecksum(buf[i:i+16], table) {
					log.Println("CRC failed.")
					checkGT = true
					continue
				}
				log.Println("CRC passed!")
				copy(cleanBuf[j:j+12], buf[i:i+12])
				j += 12
			}
			// add extra bytes (if any) to incompleteBuf
			copy(incompleteBuf[:n-i+1], buf[i:n])
			bytesRead = n - i + 1
			// write complete frames to TCP
			_, err = conn.Write(cleanBuf[:j])
			if err != nil {
				log.Fatalf("Error writing to connection: %s", err)
			}
		}
	}
	// otherwise, receive messages from serial port
	for {
		n, err := s.Read(buf)
		if err != nil {
			log.Fatalf("Serial error: %s", err)
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
