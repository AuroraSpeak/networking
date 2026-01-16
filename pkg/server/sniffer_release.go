//go:build !debug
// +build !debug

package server

import "net"

type SnifferDirection string

const (
	SnifferIn  SnifferDirection = "in"
	SnifferOut SnifferDirection = "out"
)

type SnifferCallback func(dir SnifferDirection, local, remote string, payload []byte, clientID int, packetType string)

func SetSnifferCallback(cb SnifferCallback) {}

func (s *Server) capturePacket(dir SnifferDirection, remote *net.UDPAddr, payload []byte, clientID int, packetType string) {
}
