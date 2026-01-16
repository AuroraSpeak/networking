package orchestrator

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/aura-speak/networking/pkg/client"
	"github.com/aura-speak/networking/pkg/protocol"
	log "github.com/sirupsen/logrus"
)

// ClientFactory creates UDP clients (allows for debug/release builds)
type ClientFactory func(host string, port int, id int) *client.Client

// UDPClientService manages multiple UDP clients
type UDPClientService struct {
	clients          map[string]UDPClient
	clientCommandChs map[int]chan client.InternalCommand
	mu               sync.Mutex
	ctx              context.Context
	port             int
	idCounter        *IDCounter
	clientFactory    ClientFactory
}

// NewUDPClientService creates a new UDP client service
func NewUDPClientService(ctx context.Context, port int, factory ClientFactory) *UDPClientService {
	return &UDPClientService{
		clients:          make(map[string]UDPClient),
		clientCommandChs: make(map[int]chan client.InternalCommand),
		ctx:              ctx,
		port:             port,
		idCounter:        NewIDCounter(),
		clientFactory:    factory,
	}
}

// CreateClient creates a new UDP client with a random name
func (s *UDPClientService) CreateClient(onPacket func(name string, payload []byte) error) (string, int, *client.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := GetRandomName()
	// Ensure unique name
	for _, exists := s.clients[name]; exists; _, exists = s.clients[name] {
		name = GetRandomName()
	}

	id := s.idCounter.GetNextID()
	c := s.clientFactory("localhost", s.port, id)

	s.clients[name] = UDPClient{
		ID:        id,
		Client:    c,
		Name:      name,
		Datagrams: []Datagram{},
		Running:   false,
	}

	// Register packet handler
	c.OnPacket(protocol.PacketTypeDebugAny, func(packet *protocol.Packet) error {
		return onPacket(name, packet.Payload)
	})

	// Store command channel
	s.clientCommandChs[id] = c.OutCommandCh

	log.WithField("caller", "orchestrator").Infof("UDP client created: %s with id %d", name, id)
	return name, id, c, nil
}

// StopClient stops a client by name
func (s *UDPClientService) StopClient(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	uc, ok := s.clients[name]
	if !ok {
		return fmt.Errorf("UDP client not found: %s", name)
	}

	uc.Client.Stop()
	delete(s.clientCommandChs, uc.ID)
	return nil
}

// GetClient returns a client by name
func (s *UDPClientService) GetClient(name string) (UDPClient, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	uc, ok := s.clients[name]
	return uc, ok
}

// GetClientByID returns a client by ID
func (s *UDPClientService) GetClientByID(id int) (UDPClient, string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for name, uc := range s.clients {
		if uc.ID == id {
			return uc, name, true
		}
	}
	return UDPClient{}, "", false
}

// GetAllClients returns all clients
func (s *UDPClientService) GetAllClients() map[string]UDPClient {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string]UDPClient, len(s.clients))
	for k, v := range s.clients {
		result[k] = v
	}
	return result
}

// UpdateClientRunning updates the running state of a client
func (s *UDPClientService) UpdateClientRunning(id int, running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, uc := range s.clients {
		if uc.ID == id {
			uc.Running = running
			s.clients[name] = uc
			return
		}
	}
}

// AddDatagram adds a datagram to a client's history
func (s *UDPClientService) AddDatagram(name string, direction DatagramDirection, message []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	uc, ok := s.clients[name]
	if !ok {
		return fmt.Errorf("UDP client not found: %s", name)
	}

	uc.Datagrams = append(uc.Datagrams, Datagram{
		Direction: direction,
		Message:   message,
	})
	s.clients[name] = uc
	return nil
}

// GetCommandChannel returns the command channel for a client
func (s *UDPClientService) GetCommandChannel(id int) (chan client.InternalCommand, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch, ok := s.clientCommandChs[id]
	return ch, ok
}

// SendDatagram sends a datagram through a client
func (s *UDPClientService) SendDatagram(id int, message string, format string) error {
	s.mu.Lock()
	uc, name, ok := s.findClientByIDLocked(id)
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("UDP client not found")
	}

	if !uc.Running {
		return fmt.Errorf("client is not running")
	}

	messageBytes, err := ConvertMessageToBytes(message, format)
	if err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	packet := &protocol.Packet{
		PacketHeader: protocol.Header{PacketType: protocol.PacketTypeDebugAny},
		Payload:      messageBytes,
	}

	if err := uc.Client.Send(packet.Encode()); err != nil {
		return fmt.Errorf("failed to send: %w", err)
	}

	// Record datagram
	s.AddDatagram(name, ClientToServer, messageBytes)
	return nil
}

func (s *UDPClientService) findClientByIDLocked(id int) (UDPClient, string, bool) {
	for name, uc := range s.clients {
		if uc.ID == id {
			return uc, name, true
		}
	}
	return UDPClient{}, "", false
}

// ConvertMessageToBytes converts a message string to bytes based on format
func ConvertMessageToBytes(message string, format string) ([]byte, error) {
	if format == "hex" {
		hexString := strings.ReplaceAll(strings.TrimSpace(message), " ", "")
		return hex.DecodeString(hexString)
	}
	return []byte(message), nil
}
