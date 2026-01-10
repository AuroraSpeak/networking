package web

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	webutil "github.com/aura-speak/networking/internal/web/utils"
	"github.com/aura-speak/networking/pkg/client"
	log "github.com/sirupsen/logrus"
)

// genUDPClient creates a new UDP client and returns its name
func (s *Server) genUDPClient(port int) string {
	name := webutil.GetFirstName()
	id := idCounter.getNextID()
	client := client.NewDebugClient("localhost", port, id)
	s.udpClients[name] = udpClient{
		id:        id,
		client:    client,
		name:      name,
		datagrams: []datagram{},
	}
	// Register client command channel and start listening
	s.clientCommandChs[id] = client.OutCommandCh
	s.handleClientCommands(id, client.OutCommandCh)
	log.Infof("UDP client started: %s with id %d", name, id)
	return name
}

// convertMessageToBytes converts a message string to []byte based on format
func convertMessageToBytes(message string, format string) ([]byte, error) {
	if format == "hex" {
		// Remove spaces from hex string
		hexString := strings.ReplaceAll(strings.TrimSpace(message), " ", "")
		return hex.DecodeString(hexString)
	}
	// Text format: convert string to []byte using UTF-8
	return []byte(message), nil
}

// handleAllClient handles all incoming packets from UDP clients
func (s *Server) handleAllClient(name string, packet []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	udpClient, ok := s.udpClients[name]
	if !ok {
		log.Errorf("UDP client not found: %s", name)
		return fmt.Errorf("UDP client not found: %s", name)
	}

	newDatagram := datagram{
		Direction: ServerToClient,
		Message:   packet,
	}
	udpClient.datagrams = append(udpClient.datagrams, newDatagram)
	s.udpClients[name] = udpClient
	fmt.Println("handleAllClient", "broadcasting usu", udpClient.id)
	if s.wsHub != nil {
		s.wsHub.Broadcast([]byte("usu" + strconv.Itoa(udpClient.id)))
	}

	return nil
}
