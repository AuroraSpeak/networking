package dtls

import (
	"net"
	"sync"
	"time"
)

type Server struct {
	config Config

	conn  *net.UDPConn
	mu    sync.Mutex
	conns map[string]*Conn
}

func cloneUDPAddr(a *net.UDPAddr) *net.UDPAddr {
	ip := append([]byte(nil), a.IP...)
	return &net.UDPAddr{IP: ip, Port: a.Port, Zone: a.Zone}
}

func (s *Server) lookupOrCreatePeer(addr *net.UDPAddr) *Conn {
	key := addr.String()

	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.conns[key]; ok {
		return c
	}

	c := &Conn{
		UDP:          s.conn,
		Addr:         cloneUDPAddr(addr),
		inbox:        make(chan inboundDatagram),
		readTimeout:  s.config.Timeouts.ReadTimeout,
		writeTimeout: s.config.Timeouts.WriteTimeout,
		idleTimeout:  s.config.Timeouts.IdleTimeout,
		appTx:        make(chan []byte),
		appRx:        make(chan []byte),
		mtu:          s.config.MTU,
		closeCh:      make(chan struct{}),
		idleTimer:    nil,
		idleTimerMu:  sync.Mutex{},
	}
	s.conns[key] = c

	go connLoop(c)
	return c
}

func (s *Server) Serve() error {
	buf := make([]byte, 64*1024)

	for {
		n, addr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		pkg := append([]byte(nil), buf[:n]...)

		c := s.lookupOrCreatePeer(addr)
		select {
		case c.inbox <- inboundDatagram{data: pkg, addr: addr, ts: time.Now()}:
		default:
		}
	}
}
