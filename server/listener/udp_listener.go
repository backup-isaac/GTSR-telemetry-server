package listener

import (
	"log"
	"net"
	"server/configs"
)

// UDPHandler is the object representing the UDP listener
type UDPHandler struct {
	Publisher DatapointPublisher
	Parser    PacketParser
}

// NewUDPHandler returns an initialized UDPHandler
func NewUDPHandler(publisher DatapointPublisher, parser PacketParser) *UDPHandler {
	return &UDPHandler{
		Publisher: publisher,
		Parser:    parser,
	}
}

// UDPListen listens for UDP data
// This code runs in a single goroutine since UDP is connectionless
// Read data gets streamed to the DatapointPublisher
func UDPListen() {
	canConfigs, err := configs.LoadConfigs()

	if err != nil {
		log.Fatalf("Error loading CAN configs: %s", err)
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: connPort,
		IP:   net.ParseIP(connHost),
	})

	if err != nil {
		log.Fatalf("Error listening on UDP port: %s", err)
	}
	defer conn.Close()
	log.Printf("Listening on %s:%d UDP\n", connHost, connPort)

	// initialize the UDP handler to handle all UDP packets
	handler := NewUDPHandler(GetDatapointPublisher(), NewPacketParser(canConfigs))

	for {
		message := make([]byte, 20)
		rlen, _, err := conn.ReadFromUDP(message[:])

		if err != nil {
			log.Fatalf("UDP error: %s", err)
		}

		for i := 0; i < rlen; i++ {
			if handler.Parser.ParseByte(message[i]) {
				points := handler.Parser.ParsePacket()
				for _, point := range points {
					handler.Publisher.Publish(point)
				}
			}
		}

	}
}
