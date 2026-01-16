package sniffer

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Service manages packet collection and storage
type Service struct {
	packets  []Packet
	packetMu sync.Mutex
	ctx      context.Context
}

// NewService creates a new sniffer service
func NewService(ctx context.Context) *Service {
	return &Service{
		packets: make([]Packet, 0),
		ctx:     ctx,
	}
}

// Add appends a packet to the collection
// Example:
//
//	packet := Packet{
//		TS:         time.Now(),
//		Dir:        In,
//		Local:      "192.168.1.1",
//		Remote:     "192.168.1.2",
//		Payload:    []byte{0x01, 0x02, 0x03, 0x04},
//		ClientID:   1,
//		PacketType: "dtls",
//	}
//	s.Add(packet)
func (s *Service) Add(packet Packet) {
	s.packetMu.Lock()
	defer s.packetMu.Unlock()
	s.packets = append(s.packets, packet)
	log.WithFields(log.Fields{
		"caller": "sniffer",
		"cid":    packet.ClientID,
		"dir":    packet.Dir,
		"len":    len(packet.Payload),
	}).Debugf("Captured packet: %s -> %s", packet.Local, packet.Remote)
}

// GetAll returns all captured packets
// Example:
//
//	packets := s.GetAll()
//	for _, packet := range packets {
//		fmt.Println(packet)
//	}
func (s *Service) GetAll() []Packet {
	s.packetMu.Lock()
	defer s.packetMu.Unlock()

	result := make([]Packet, len(s.packets))
	copy(result, s.packets)
	return result
}

// GetByClientID returns all packets for a specific client
// If clientID is 0, returns all packets without a client assigned
// Example:
//
//	packets := s.GetByClientID(1)
//	for _, packet := range packets {
//		fmt.Println(packet)
//	}
func (s *Service) GetByClientID(clientID int) []Packet {
	s.packetMu.Lock()
	defer s.packetMu.Unlock()

	filtered := make([]Packet, 0)
	for _, p := range s.packets {
		if p.ClientID == clientID {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// GetUnassigned returns all packets without an assigned client (ClientID = 0)
// Example:
//
//	packets := s.GetUnassigned()
//	for _, packet := range packets {
//		fmt.Println(packet)
//	}
func (s *Service) GetUnassigned() []Packet {
	return s.GetByClientID(0)
}

// Clear removes all captured packets
// Example:
//
//	s.Clear()
func (s *Service) Clear() {
	s.packetMu.Lock()
	defer s.packetMu.Unlock()
	s.packets = make([]Packet, 0)
}
