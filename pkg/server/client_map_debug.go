//go:build debug
// +build debug

package server

import (
	"sync"

	"github.com/aura-speak/networking/pkg/protocol"
	log "github.com/sirupsen/logrus"
)

var dbgAddrToClientMap = sync.Map{}

func tryRegisterClient(remote string, id int) (registered bool) {
	dbgAddrToClientMap.Store(remote, id)
	return true
}

func lookupClientID(remote string) (int, bool) {
	v, ok := dbgAddrToClientMap.Load(remote)
	if !ok {
		return 0, false
	}
	id, ok := v.(int)
	return id, ok
}

func (s *Server) handleDebugHello(packet *protocol.Packet, clientAddr string) error {
	log.WithField("caller", "server").Infof("Received debug hello packet from %s: %d", clientAddr, packet.Payload[0])
	tryRegisterClient(clientAddr, int(packet.Payload[0]))
	return nil
}
