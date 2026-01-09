//go:build debug
// +build debug

package client

import (
	"context"
	"sync/atomic"

	"github.com/aura-speak/networking/pkg/router"
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
		packetRouter: router.NewPacketRouter(),
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
