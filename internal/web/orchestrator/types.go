package orchestrator

import (
	"github.com/aura-speak/networking/pkg/client"
)

// DatagramDirection indicates the direction of a datagram
type DatagramDirection int

const (
	ClientToServer DatagramDirection = 1
	ServerToClient DatagramDirection = 2
)

// Datagram represents a single UDP datagram
type Datagram struct {
	Direction DatagramDirection `json:"direction"`
	Message   []byte            `json:"message"`
}

// UDPClient wraps a client with additional metadata
type UDPClient struct {
	ID        int
	Client    *client.Client
	Name      string
	Datagrams []Datagram
	Running   bool
}

// UDPClientAction defines client actions
type UDPClientAction int

const (
	UDPClientActionAddDatagram UDPClientAction = iota
)

// UDPClientActionData carries action data for a client
type UDPClientActionData struct {
	ID     int
	Action UDPClientAction
}

// Caller identifies the source of internal messages
type Caller string

const (
	CallerUDPClient Caller = "UDPClient"
	CallerUDPServer Caller = "UDPServer"
	CallerWebServer Caller = "WebServer"
)

// InternalMessage is used for internal communication
type InternalMessage struct {
	Caller  Caller `json:"caller"`
	Target  string `json:"target"`
	Content string `json:"content"`
}

// ToBytes converts the message content to bytes
func (m *InternalMessage) ToBytes() []byte {
	return []byte(m.Content)
}

// FromBytes sets the content from bytes
func (m *InternalMessage) FromBytes(data []byte) {
	m.Content = string(data)
}

// ToJSON returns a JSON representation
func (m *InternalMessage) ToJSON() string {
	return `{"caller":"` + string(m.Caller) + `","target":"` + m.Target + `","content":"` + m.Content + `"}`
}
