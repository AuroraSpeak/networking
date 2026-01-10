//go:build !debug
// +build !debug

package client

import log "github.com/sirupsen/logrus"

func NewDebugClient(Host string, Port int, ID int) *Client {
	log.Error("NewDebugClient is not implemented in release build")
	return nil
}

func (c *Client) debugHello() {}
