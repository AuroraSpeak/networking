//go:build !debug
// +build !debug

package server

import "net"

type TraceDirection string

const (
	TraceIn  TraceDirection = "in"
	TraceOut TraceDirection = "out"
)

type TraceEvent struct{}

func (s *Server) initTracer() chan TraceEvent { return nil }

func NewTraceEvent(dir TraceDirection, local string, remote string, len int, payload []byte, clientID int) *TraceEvent {
	return nil
}

func (s *Server) trace(dir TraceDirection, remote *net.UDPAddr, payload []byte) {}
