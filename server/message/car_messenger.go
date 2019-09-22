package message

import "server/listener"

// CarMessenger handles sending messages to the car
type CarMessenger struct {
	TCPPrefix string
}

// NewCarMessenger returns a new Messenger initialized with the provided TCP
// prefix, and a message terminator used when constructing a TCP message
func NewCarMessenger(tcpPrefix string) *CarMessenger {
	return &CarMessenger{TCPPrefix: tcpPrefix}
}

// UploadTCPMessage sends the provided message to the listener, which will then
// relay it to the car
func (m *CarMessenger) UploadTCPMessage(message string) {
	msg := m.TCPPrefix + string(len(message)) + message
	listener.Write([]byte(msg))
}
