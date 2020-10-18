package websocket

import "math/rand"

// Message is both requests and responses are sent as packets wia WebSocket.
// Their payload follows the following basic structure.
type Message struct {
	// Message field is a string command to send to the server and response
	// message from RemoteServer.
	Message string `json:"Message"`

	// Identifier field is a 32-bit little endian integer chosen by the
	// client for each request. It may be set to any positive integer.
	// When the RemoteServer responds to the request, the response packet
	// will have the same Identifier as the original request.
	// It need not be unique, but if a unique packet id is assigned,
	// it can be used to match incoming responses to their corresponding requests.
	Identifier int `json:"Identifier"`

	// Type is the type of message that was sent or received.
	// Can take the following values: Generic, Log, Warning, Error.
	// When sending a request, you can leave it blank.
	Type string `json:"Type"`

	Stacktrace string `json:"stacktrace"`
}

func newMessage(command string) *Message {
	return &Message{
		Message:    command,
		Identifier: rand.Intn(1000),
	}
}
