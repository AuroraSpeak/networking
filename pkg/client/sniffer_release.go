//go:build !debug
// +build !debug

package client

type ClientSnifferDirection string

const (
	ClientSnifferIn  ClientSnifferDirection = "in"
	ClientSnifferOut ClientSnifferDirection = "out"
)

type ClientSnifferCallback func(dir ClientSnifferDirection, local, remote string, payload []byte, clientID int, packetType string)

func SetClientSnifferCallback(cb ClientSnifferCallback) {}

func (c *Client) capturePacket(dir ClientSnifferDirection, payload []byte, clientID int, packetType string) {
}
