package orchestrator

import (
	"context"
	"sync"

	"github.com/aura-speak/networking/pkg/protocol"
	"github.com/aura-speak/networking/pkg/server"
	log "github.com/sirupsen/logrus"
)

// UDPServerService manages the UDP server lifecycle
type UDPServerService struct {
	server *server.Server
	mu     sync.Mutex
	ctx    context.Context
	port   int
}

// NewUDPServerService creates a new UDP server service
func NewUDPServerService(ctx context.Context, port int) *UDPServerService {
	return &UDPServerService{
		ctx:  ctx,
		port: port,
	}
}

// Start creates and starts the UDP server
func (s *UDPServerService) Start(onPacket func(clientAddr string, payload []byte) error) (*server.Server, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.server = server.NewServer(s.port, s.ctx)
	s.server.OnPacket(protocol.PacketTypeDebugAny, func(packet *protocol.Packet, clientAddr string) error {
		return onPacket(clientAddr, packet.Payload)
	})

	return s.server, nil
}

// Stop stops the UDP server
func (s *UDPServerService) Stop() error {
	s.mu.Lock()
	srv := s.server
	s.mu.Unlock()

	if srv == nil {
		return nil
	}

	srv.Stop()
	log.WithField("caller", "orchestrator").Info("UDP server stopped")
	return nil
}

// GetServer returns the current server instance
func (s *UDPServerService) GetServer() *server.Server {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.server
}

// IsRunning returns whether the server is currently running
func (s *UDPServerService) IsRunning() bool {
	s.mu.Lock()
	srv := s.server
	s.mu.Unlock()

	if srv == nil {
		return false
	}
	return srv.ServerState.IsAlive
}

// GetState returns the current server state
func (s *UDPServerService) GetState() (shouldStop bool, isAlive bool) {
	s.mu.Lock()
	srv := s.server
	s.mu.Unlock()

	if srv == nil {
		return false, false
	}
	return srv.ServerState.ShouldStop, srv.ServerState.IsAlive
}
