package message

// Writer handles writing a message to a TCP listener
type Writer interface {
	Write([]byte)
}

// CarMessenger handles sending messages to the car
type CarMessenger struct {
	TCPPrefix string
	Writer    Writer
}

// NewCarMessenger returns a new Messenger initialized with the provided TCP
// prefix that will write new messages to the provided Writer
func NewCarMessenger(tcpPrefix string, writer Writer) *CarMessenger {
	return &CarMessenger{
		TCPPrefix: tcpPrefix,
		Writer:    writer,
	}
}

// UploadTCPMessage sends the provided message to the listener, which will then
// relay it to the car
func (m *CarMessenger) UploadTCPMessage(message string) {
	constructedMsg := make([]byte, 0)

	constructedMsg = append(constructedMsg, []byte(m.TCPPrefix)...)
	constructedMsg = append(constructedMsg, byte(len(message)))
	constructedMsg = append(constructedMsg, []byte(message)...)

	m.Writer.Write(constructedMsg)
}
