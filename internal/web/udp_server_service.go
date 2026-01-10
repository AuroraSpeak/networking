package web

import (
	"github.com/aura-speak/networking/pkg/protocol"
	log "github.com/sirupsen/logrus"
)

// handleAll handles all incoming packets from UDP server
func (s *Server) handleAll(clientAddr string, packet []byte) error {
	log.WithField("caller", "web").Infof("Received packet from %s: %s", clientAddr, string(packet))
	s.mu.Lock()
	if s.udpServer != nil {
		s.udpServer.Broadcast(&protocol.Packet{
			PacketHeader: protocol.Header{PacketType: protocol.PacketTypeDebugAny},
			Payload:      packet,
		})
	}
	s.mu.Unlock()
	s.mu.Lock()
	if s.wsHub != nil {
		s.wsHub.Broadcast(packet)
	}
	s.mu.Unlock()
	return nil
}
