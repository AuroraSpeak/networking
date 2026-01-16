//go:build debug
// +build debug

package server

import (
	"net"
)

// SnifferCallback is a function type for capturing packets
type SnifferCallback func(dir SnifferDirection, local, remote string, payload []byte, clientID int, packetType string)

type SnifferDirection string

const (
	SnifferIn  SnifferDirection = "in"
	SnifferOut SnifferDirection = "out"
)

var snifferCallback SnifferCallback

// SetSnifferCallback sets the callback function for packet sniffing
func SetSnifferCallback(cb SnifferCallback) {
	snifferCallback = cb
}

// capturePacket captures a packet if the callback is set
func (s *Server) capturePacket(dir SnifferDirection, remote *net.UDPAddr, payload []byte, clientID int, packetType string) {
	if snifferCallback == nil {
		return
	}

	local := ""
	remoteAddr := ""
	if s.conn != nil && s.conn.LocalAddr() != nil {
		local = s.conn.LocalAddr().String()
	}
	if remote != nil {
		remoteAddr = remote.String()
	}
	if local == "" {
		local = "unknown"
	}
	if remoteAddr == "" {
		remoteAddr = "unknown"
	}

	snifferCallback(dir, local, remoteAddr, payload, clientID, packetType)
}
