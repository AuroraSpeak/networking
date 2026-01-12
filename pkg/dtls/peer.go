package dtls

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

type Conn struct {
	UDP  *net.UDPConn
	Addr *net.UDPAddr

	inbox chan inboundDatagram

	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration

	appRx   chan []byte
	appTx   chan []byte
	mtu     int
	closeCh chan struct{}

	idleTimer   *time.Timer
	idleTimerMu sync.Mutex
}

func (c *Conn) Send(b []byte) (int, error) {
	if len(b) > c.mtu {
		return 0, errors.New("datagram too long")
	}
	if c.writeTimeout <= 0 {
		n, err := c.UDP.WriteToUDP(b, c.Addr)
		if err == nil {
			c.resetIdleTimer()
		}
		return n, err
	}

	timer := time.NewTimer(c.writeTimeout)
	defer timer.Stop()
	select {
	case c.appTx <- b:
		c.resetIdleTimer()
		return len(b), nil
	case <-timer.C:
		return 0, io.EOF
	case <-c.closeCh:
		return 0, io.EOF
	}
}

func (c *Conn) Read(p []byte) (int, error) {
	if c.readTimeout <= 0 {
		b, ok := <-c.appRx
		if !ok {
			return 0, io.EOF
		}
		return copy(p, b), nil
	}

	timer := time.NewTimer(c.readTimeout)
	defer timer.Stop()

	select {
	case b, ok := <-c.appRx:
		if !ok {
			return 0, io.EOF
		}
		return copy(p, b), nil
	case <-timer.C:
		return 0, io.EOF
	case <-c.closeCh:
		return 0, io.EOF
	}
}

func connLoop(c *Conn) {
	var idleTimerCh <-chan time.Time

	c.idleTimerMu.Lock()
	if c.idleTimeout > 0 {
		c.idleTimer = time.NewTimer(c.idleTimeout)
		idleTimerCh = c.idleTimer.C
	}
	c.idleTimerMu.Unlock()

	if c.idleTimer != nil {
		defer c.idleTimer.Stop()
	}

	for {
		select {
		case d, ok := <-c.inbox:
			if !ok {
				return
			}
			c.resetIdleTimer()
			c.handleDatagram(d)
		case <-idleTimerCh:
			// idleTimerCh ist nil wenn idleTimeout <= 0, dann wird dieser case nie ausgewählt
			close(c.closeCh)
			return
		case <-c.closeCh:
			return
		}
	}
}

func (c *Conn) handleDatagram(d inboundDatagram) {
}

func (c *Conn) resetIdleTimer() {
	if c.idleTimeout <= 0 {
		return
	}
	c.idleTimerMu.Lock()
	defer c.idleTimerMu.Unlock()

	if c.idleTimer != nil {
		if !c.idleTimer.Stop() {
			select {
			case <-c.idleTimer.C:
			default:
			}
		}
		c.idleTimer.Reset(c.idleTimeout)
	}
}
