//go:build debug
// +build debug

package client

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/aura-speak/networking/pkg/protocol"
	"github.com/aura-speak/networking/pkg/router"
	log "github.com/sirupsen/logrus"
)

func NewDebugClient(Host string, Port int, ID int) *Client {
	ctx := context.Background()
	c := &Client{
		Host:         Host,
		Port:         Port,
		sendCh:       make(chan []byte),
		recvCh:       make(chan []byte),
		errCh:        make(chan error),
		ctx:          ctx,
		packetRouter: router.NewClientPacketRouter(),
		ClientState: ClientState{
			ID: ID,
		},
		OutCommandCh: make(chan InternalCommand, 10),
	}
	return c
}

func (c *Client) SetRunningState(running bool) {
	var v int32
	if running {
		v = 1
	}
	atomic.StoreInt32(&c.ClientState.Running, v)

	// Notify web server about state change
	select {
	case <-c.ctx.Done():
		return
	case c.OutCommandCh <- CmdUpdateClientState:
	default:
		// Channel full, skip notification (non-blocking)
	}
}

func (c *Client) debugHello() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for c.conn == nil {
		select {
		case <-ctx.Done():
			log.WithField("caller", "client").Warn("Timeout waiting for connection in debugHello")
			return
		case <-time.After(10 * time.Millisecond):
			// Kurz warten und erneut prüfen
		}
	}
	packet := &protocol.Packet{
		PacketHeader: protocol.Header{PacketType: protocol.PacketTypeDebugHello},
		Payload:      []byte(strconv.Itoa(c.ClientState.ID)),
	}
	log.WithField("caller", "client").Infof("Sending debug hello packet to %s: %d", c.conn.RemoteAddr().String(), c.ClientState.ID)
	c.Send(packet.Encode())
}
