package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/tarm/serial"
)

func main() {
	var host string
	var serialPort string

	// connect to the serial port
	if len(os.Args) > 1 {
		serialPort = os.Args[1]
	} else {
		// TODO: Mock Serial Connection?
		log.Fatalf("Please specify serial port (e.g. COM4 or /dev/ttyUSB0)")
	}

	c := &serial.Config{Name: serialPort, Baud: 115200}
	s, err := serial.OpenPort(c)
	defer s.Close()
	if err != nil {
		log.Fatalf("Serial Error: %s", err)
	} else {
		log.Printf("Successfully connected to %s\n", serialPort)
	}

	// set if we are uploading to the production server or localhost
	if len(os.Args) > 2 && os.Args[2] == "remote" {
		host = "solarracing.me"
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
	go listen(conn)

	// receive messages from serial port
	buf := make([]byte, 128)
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
// TODO: Relay via Serial back to the car
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
