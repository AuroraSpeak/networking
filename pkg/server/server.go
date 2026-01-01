package server

import (
	"context"
	"net"
	"sync"
)

type Server struct {
	conn        *net.UDPConn
	Port        int
	remoteConns *sync.Map
	ctx         context.Context

	// stopping sign for the Run loop
	shouldStop bool

	// incoming message channel
	IncomingCh chan []byte
}

func NewServer(port int, ctx context.Context) *Server {
	return &Server{
		Port:        port,
		remoteConns: new(sync.Map),
		ctx:         ctx,
	}
}

func (s *Server) Run() error {
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

	// Infinite loop that listens for incoming UDP packets
	for !s.shouldStop {

		// ! TMP: Change this Later
		// Buffer to hold incoming data
		buf := make([]byte, 1024)
		_, remoteAddr, err := s.conn.ReadFrom(buf)
		if err != nil {
			continue
		}

		// Check if it is a new remote address
		// If so, store it in the map
		if _, ok := s.remoteConns.Load(remoteAddr.String()); !ok {
			s.remoteConns.Store(remoteAddr.String(), &remoteAddr)
		}
		s.IncomingCh <- buf

		// Broadcast the received packet to all connected clients
		go func() {
			s.remoteConns.Range(func(key, value any) bool {
				if _, err := s.conn.WriteTo(buf, *value.(*net.Addr)); err != nil {
					// Remove client if needed
					s.remoteConns.Delete(key)
					return true
				}
				return true
			})
		}()
	}
	return nil
}

func (s *Server) Broadcast(message []byte) {
	s.remoteConns.Range(func(key, value any) bool {
		if _, err := s.conn.WriteTo(message, *value.(*net.Addr)); err != nil {
			// Remove client if needed
			s.remoteConns.Delete(key)
			return true
		}
		return true
	})
}

func (s *Server) Stop() {
	s.shouldStop = true
	
	// Send stop message to all connected clients
	if s.conn != nil {
		s.remoteConns.Range(func(key, value any) bool {
			s.conn.WriteTo([]byte("STOP"), *value.(*net.Addr))
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
