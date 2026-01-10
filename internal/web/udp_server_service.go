package web

import (
	log "github.com/sirupsen/logrus"
)

// handleAll handles all incoming packets from UDP server
func (s *Server) handleAll(packet []byte) error {
	log.Infof("Received packet: %s", string(packet))
	s.mu.Lock()
	if s.udpServer != nil {
		s.udpServer.Broadcast(packet)
	}
	s.mu.Unlock()
	s.mu.Lock()
	if s.wsHub != nil {
		s.wsHub.Broadcast(packet)
	}
	s.mu.Unlock()
	return nil
}
