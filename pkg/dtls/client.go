package dtls

import (
	"context"
	"net"
	"sync"
)

type Client struct {
	config Config

	// Here is the Send Function in
	Conn
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (c *Client) resolveUDPAddr(raddr string, port string) error {
	mapping := net.JoinHostPort(raddr, port)
	addr, err := net.ResolveUDPAddr("udp", mapping)
	if err != nil {
		return err
	}
	c.Addr = addr
	return nil
}

func (c *Client) DialWithPort(raddr, port string) error {
	err := c.resolveUDPAddr(raddr, port)
	if err != nil {
		return err
	}

	return c.dial()
}

func (c *Client) Dial(raddr string) error {
	err := c.resolveUDPAddr(raddr, DEFAULT_PORT)
	if err != nil {
		return err
	}

	return c.dial()
}

func (c *Client) dial() error {
	c.ctx, c.cancel = context.WithCancel(context.Background())
	conn, err := net.DialUDP("udp", nil, c.Addr)
	if err != nil {
		return err
	}
	c.UDP = conn

	c.wg.Go(func() {
		connLoop(&Conn{
			UDP:          conn,
			Addr:         c.Addr,
			inbox:        c.inbox,
			readTimeout:  c.readTimeout,
			writeTimeout: c.writeTimeout,
			idleTimeout:  c.idleTimeout,
			appTx:        c.appTx,
			appRx:        c.appRx,
			mtu:          c.mtu,
			closeCh:      c.closeCh,
			idleTimer:    c.idleTimer,
			idleTimerMu:  sync.Mutex{},
		})
	})
	return nil
}

func (c *Client) close() {
	c.cancel()
	c.wg.Wait()
	c.UDP.Close()
}
