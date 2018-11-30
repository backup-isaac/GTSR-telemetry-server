package main

import (
	"fmt"
	"log"
	"net"
	"os"

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
<<<<<<< Updated upstream
func listen(conn net.Conn) {
=======
func listen(conn net.Conn, s serial.Port) {
>>>>>>> Stashed changes
	buf := make([]byte, 128)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("Error reading from connection: %s", err)
		}
		log.Printf("Received message from server: %q", buf[:n])
	}
}
