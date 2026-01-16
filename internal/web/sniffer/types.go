package sniffer

import "time"

type Direction string

const (
	In  Direction = "in"
	Out Direction = "out"
)

// Packet represents a captured network packet
type Packet struct {
	TS         time.Time `json:"ts"`
	Dir        Direction `json:"dir"`
	Local      string    `json:"local"`
	Remote     string    `json:"remote"`
	Payload    []byte    `json:"payload"`
	ClientID   int       `json:"client_id"`             // 0 = unbekannt/kein Client zugeordnet
	PacketType string    `json:"packet_type,omitempty"` // Optional, leer wenn nicht dekodierbar
}

// NewPacket creates a new packet with current timestamp
func NewPacket(dir Direction, local, remote string, payload []byte, clientID int, packetType string) Packet {
	return Packet{
		TS:         time.Now(),
		Dir:        dir,
		Local:      local,
		Remote:     remote,
		Payload:    payload,
		ClientID:   clientID,
		PacketType: packetType,
	}
}
