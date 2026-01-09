//go:build !debug
// +build !debug

package client

import "sync/atomic"

func (c *Client) SetRunningState(running bool) {
	var v int32
	if running {
		v = 1
	}
	atomic.StoreInt32(&c.ClientState.Running, v)
	// OutCommandCh is nil in non-debug builds, so no notification is sent
}
