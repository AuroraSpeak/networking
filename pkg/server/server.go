package server

import (
	"net"
	"sync"
)

type Server struct {
	addr        net.UDPAddr
	Port        string
	remoteConns *sync.Map
}

func NewServer(port string) *Server {
	return &Server{
		Port:        port,
		remoteConns: new(sync.Map),
	}
}

func (s *Server) Run() error {
	conn, err := net.ListenUDP("udp", &s.addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Infinite loop that listens for incoming UDP packets
	for {
		// ! TMP: Change this Later
		// Buffer to hold incoming data
		buf := make([]byte, 1024)
		_, remoteAddr, err := conn.ReadFrom(buf)
		if err != nil {
			continue
		}

		// Check if it is a new remote address
		// If so, store it in the map
		if _, ok := s.remoteConns.Load(remoteAddr.String()); !ok {
			s.remoteConns.Store(remoteAddr.String(), &remoteAddr)
		}

		// Broadcast the received packet to all connected clients
		go func() {
			s.remoteConns.Range(func(key, value any) bool {
				if _, err := conn.WriteTo(buf, *value.(*net.Addr)); err != nil {
					// Remove client if needed
					s.remoteConns.Delete(key)
					return true
				}
				return true
			})
		}()
	}
}
