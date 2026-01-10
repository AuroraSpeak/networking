// Package Client contains the core networking for the Client
// It is responsible for sending and receiving packets to the Server
// The implementation is based on the UDP protocol
// It Handles Client States, Starts and Runs the Client
// Add callbacks to the Client to handle incoming packets
// It may contains some default callbacks required for handling the client state
package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/aura-speak/networking/pkg/router"
	log "github.com/sirupsen/logrus"
)

// Gedanken:
// 	- ein isAlive wie im server, aber dies wird nachher über das protokoll gehändelt

// Client is the main struct for the UDP Client
// It contains the connection to the UDP Client
// The Host of the Client
// The Port of the Client
// The context of the Client
// The cancel function for the Client
// The wg for the Client
// The send channel for the Client
// The recv channel for the Client
// The err channel for the Client
// The packet router for the Client
// The running sign for the Client
type Client struct {
	Host string
	Port int

	// TODO: uncomment later MessageLoop

	conn *net.UDPConn

	ctx context.Context
	wg  sync.WaitGroup

	ClientState

	// Communication Channels
	// Client -> Server
	sendCh chan []byte
	// Server -> Client
	recvCh chan []byte
	// Sends errors from go Routines to main routine
	errCh chan error

	packetRouter *router.PacketRouter

	running bool

	// OutCommandCh sends internal commands to the web server (only used in debug builds)
	OutCommandCh chan InternalCommand
}

// NewClient creates a new UDP Client it takes the Host, Port and timeout of the Client
func NewClient(Host string, Port int) *Client {
	ctx := context.Background()
	c := &Client{
		Host:         Host,
		Port:         Port,
		sendCh:       make(chan []byte),
		recvCh:       make(chan []byte),
		errCh:        make(chan error),
		ctx:          ctx,
		packetRouter: router.NewPacketRouter(),
	}
	return c
}

// OnPacket registers a new PacketHandler for a specific packet type
//
// Example:
//
//	client.OnPacket("text", func(packet []byte) error {
//		fmt.Println("Received text packet:", string(packet))
//		return nil
//	})
func (c *Client) OnPacket(packetType string, handler router.PacketHandler) {
	c.packetRouter.OnPacket("", handler)
}

// Run starts the Client and connects to the Server
func (c *Client) Run() error {
	conncetionString := fmt.Sprintf("%s:%d", c.Host, c.Port)
	s, err := net.ResolveUDPAddr("udp4", conncetionString)
	if err != nil {
		return err
	}
	c.conn, err = net.DialUDP("udp", nil, s)
	if err != nil {
		return err
	}
	c.running = true
	c.SetRunningState(true)

	c.wg.Go(func() {
		c.recvLoop()
	})

	c.wg.Go(func() {
		c.sendLoop()
	})

	c.wg.Go(func() {
		c.handleErrors()
	})
	defer c.conn.Close()
	log.Info("Starting client")
	c.wg.Wait()
	c.running = false
	c.SetRunningState(false)
	log.Info("Client Stopped")
	return nil
}

// Stop stops the Client
// It stops the Client and closes the connection to the Server
func (c *Client) Stop() {
	c.running = false
	c.SetRunningState(false)
	if c.conn != nil {
		c.conn.Close()
	}
	c.wg.Wait()
	log.Info("Client Stopped")
}

// Send sends a packet to the Server
//
// Example:
//
// client.Send([]byte("Hello, Server!"))
func (c *Client) Send(msg []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-c.ctx.Done():
		return errors.New("context cancelled")
	case <-ctx.Done():
		return errors.New("send timeout: sendLoop may not be running or is blocked")
	case c.sendCh <- msg:
		return nil
	}
}

// sendLoop sends packets to the Server
func (c *Client) sendLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.sendCh:
			if _, err := c.conn.Write(msg); err != nil {
				c.errCh <- err
				continue
			}
			c.SetRunningState(true)
		}
	}
}

// recvLoop receives packets from the Server
func (c *Client) recvLoop() {
	// TODO: change later the byte size
	buffer := make([]byte, 1024)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		n, _, err := c.conn.ReadFromUDP(buffer)
		if err != nil {
			c.errCh <- err
			continue
		}
		if n == 0 {
			continue
		}
		// TODO: same here change later when we exacly know the max msg length of a packet in the Protocol
		dst := make([]byte, n)
		copy(dst, buffer[:n])
		if string(dst) == "STOP" {
			c.ctx.Done()
			log.Info("Received STOP message from server")
			return
		}

		if err := c.packetRouter.HandlePacket("", dst); err != nil {
			log.WithError(err).Error("Error handling packet")
			continue
		}
	}
}

func (c *Client) handleErrors() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case err := <-c.errCh:
			log.WithError(err).Error("Error in client")
			continue
		}
	}
}
