// Package Server contains the core networking for the Server
// It is responsible for listening for incoming packets
// The implementation is based on the UDP protocol
// It Handles Server States, Starts and Runs the Server
// Add callbacks to the Server to handle incoming packets
// It may contains some default callbacks required for handling the server state

package server

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/aura-speak/networking/pkg/protocol"
	"github.com/aura-speak/networking/pkg/router"
	log "github.com/sirupsen/logrus"
)

// NOTE: Structs

// Server is the main struct for the UDP Server
// It contains the connection to the UDP Server
// The Port of the Server
// The remote connections to the Server
// The context of the Server
// The ServerState
// The stopping sign for the Run loop
// The isAlive sign for the Server
// The shouldStop sign for the Server
// The incoming message channel
// The wg for the Server
// The out command channel
// The packet router for the Server
type Server struct {
	// Networking stuff
	Port        int
	conn        *net.UDPConn
	remoteConns *sync.Map

	ctx    context.Context
	cancel context.CancelFunc

	// ServerState: tells the state of the networking parts of the server
	ServerState

	// stopping sign for the Run loop
	IsAlive    int32
	shouldStop int32

	wg sync.WaitGroup

	// send command internal channel
	OutCommandCh chan InternalCommand

	packetRouter *router.ServerPacketRouter
	TraceCh      chan TraceEvent
}

// ServerState is the struct for the server state
// It contains the updated sign for the server state
// The shouldStop sign for the server
// The isAlive sign for the server
type ServerState struct {
	// updated says if the server Stated updated
	updated    bool `json:"-"`
	ShouldStop bool `json:"shouldStop"`
	IsAlive    bool `json:"isAlive"`
}

// NOTE: Server functions

// NewServer creates a new UDP Server it takes the port of the Server and the context of the Server
func NewServer(port int, ctx context.Context) *Server {
	childCtx, cancel := context.WithCancel(ctx)
	srv := &Server{
		Port:         port,
		remoteConns:  new(sync.Map),
		OutCommandCh: make(chan InternalCommand, 10),
		ctx:          childCtx,
		cancel:       cancel,
		packetRouter: router.NewServerPacketRouter(),
	}
	srv.initTracer()
	srv.OnPacket(protocol.PacketTypeDebugHello, srv.handleDebugHello)
	return srv
}

// OnPacket registers a new PacketHandler for a specific packet type
//
// Example:
//
//	server.OnPacket(protocol.PacketTypeDebugHello, func(packet *protocol.Packet, clientAddr string) error {
//		fmt.Println("Received text packet:", string(packet))
//		return nil
//	})
func (s *Server) OnPacket(packetType protocol.PacketType, handler router.ServerPacketHandler) {
	log.WithField("caller", "server").Infof("Registering packet handler for packet type: %s", protocol.PacketTypeMapType[packetType])
	s.packetRouter.OnPacket(packetType, handler)
}

// Run starts the Server and listens for incoming packets
func (s *Server) Run() error {
	s.packetRouter.ListRoutes()
	if atomic.LoadInt32(&s.IsAlive) == 1 {
		return errors.New("server is already running")
	}
	addr := net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: s.Port,
	}
	var err error
	s.conn, err = net.ListenUDP("udp", &addr)
	if err != nil {
		return err
	}
	defer s.conn.Close()
	s.setIsAlive(true)
	log.WithField("caller", "server").Infof("Server started on port %d", s.Port)

	// Infinite loop that listens for incoming UDP packets
	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-s.OutCommandCh:
		default:
		}
		shouldStop := atomic.LoadInt32(&s.shouldStop) == 1
		if shouldStop {
			break
		}
		// TODO: Change this Later
		// Buffer to hold incoming data
		buf := make([]byte, 1024)
		n, remoteAddr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		// Check if it is a new remote address
		// If so, store it in the map
		if _, ok := s.remoteConns.Load(remoteAddr.String()); !ok {
			s.remoteConns.Store(remoteAddr.String(), remoteAddr)
		}

		// Capture raw packet data before decoding (for DTLS and other protocols)
		rawPayload := make([]byte, n)
		copy(rawPayload, buf[:n])
		clientID, _ := lookupClientID(remoteAddr.String())
		s.capturePacket(SnifferIn, remoteAddr, rawPayload, clientID, "")

		packet, err := protocol.Decode(buf[:n])
		s.trace(TraceIn, remoteAddr, packet.Payload)
		if err != nil {
			log.WithField("caller", "server").WithError(err).Error("Error decoding packet")
			// Packet already captured above with empty packetType
			continue
		}

		// Update captured packet with packet type if decoding succeeded
		packetTypeStr := protocol.PacketTypeMapType[packet.PacketHeader.PacketType]
		if packetTypeStr != "" {
			// Re-capture with packet type (optional, for better tracking)
			s.capturePacket(SnifferIn, remoteAddr, packet.Payload, clientID, packetTypeStr)
		}

		if err := s.packetRouter.HandlePacket(packet, remoteAddr.String()); err != nil {
			log.WithField("caller", "server").WithError(err).Error("Error handling packet")
			continue
		}
	}
	s.setIsAlive(false)
	return nil
}

func (s *Server) Broadcast(packet *protocol.Packet) {
	s.wg.Go(func() {
		s.remoteConns.Range(func(key, value any) bool {
			remoteAddr := value.(*net.UDPAddr)
			encoded := packet.Encode()
			if _, err := s.conn.WriteToUDP(encoded, remoteAddr); err != nil {
				// Remove client if needed
				s.remoteConns.Delete(key)
				return true
			}
			s.trace(TraceOut, remoteAddr, packet.Payload)
			// Capture outgoing packet
			clientID, _ := lookupClientID(remoteAddr.String())
			packetTypeStr := protocol.PacketTypeMapType[packet.PacketHeader.PacketType]
			s.capturePacket(SnifferOut, remoteAddr, packet.Payload, clientID, packetTypeStr)
			return true
		})
	})
}

// Stop stops the Server and closes all connections
func (s *Server) Stop() {
	s.setShouldStop()

	// Cancel context to signal all goroutines to stop
	if s.cancel != nil {
		s.cancel()
	}

	// Send stop message to all connected clients
	if s.conn != nil {
		s.remoteConns.Range(func(key, value any) bool {
			s.conn.WriteToUDP([]byte("STOP"), value.(*net.UDPAddr))
			return true
		})

		// Close the UDP connection to interrupt ReadFrom
		s.conn.Close()
	}

	// Clear all remote connections
	s.remoteConns.Range(func(key, value any) bool {
		s.remoteConns.Delete(key)
		return true
	})
}

// setShouldStop sets the shouldStop sign for the Server
func (s *Server) setShouldStop() {
	atomic.StoreInt32(&s.shouldStop, 1)
	select {
	case <-s.ctx.Done():
		return
	case s.OutCommandCh <- CmdUpdateServerState:
	default:
	}
	s.updated = true
	s.ShouldStop = true
}

// setIsAlive sets the isAlive sign for the Server
func (s *Server) setIsAlive(val bool) {
	var v int32
	if val {
		v = 1
	}
	atomic.StoreInt32(&s.IsAlive, v)
	select {
	case <-s.ctx.Done():
		return
	case s.OutCommandCh <- CmdUpdateServerState:
	default:
	}
	s.updated = true
	s.ServerState.IsAlive = val
}
