//go:build debug
// +build debug

package server

import (
	"net"
	"time"
)

type TraceDirection string

const (
	TraceIn  TraceDirection = "in"
	TraceOut TraceDirection = "out"
)

type TraceEvent struct {
	TS       time.Time      `json:"ts"`
	Dir      TraceDirection `json:"dir"`
	Local    string         `json:"local"`
	Remote   string         `json:"remote"`
	Len      int            `json:"len"`
	Payload  []byte         `json:"payload"`
	ClientID int            `json:"client_id"`
}

func NewTraceEvent(dir TraceDirection, local string, remote string, len int, payload []byte, clientID int) TraceEvent {
	return TraceEvent{
		TS:       time.Now(),
		Dir:      dir,
		Local:    local,
		Remote:   remote,
		Len:      len,
		Payload:  payload,
		ClientID: clientID,
	}
}

func (s *Server) initTracer() chan TraceEvent {
	return make(chan TraceEvent, 2000)
}

func (s *Server) emitTrace(dir TraceDirection, local, remote string, payload []byte, clientID int) {
	if s.TraceCh == nil {
		return
	}

	if len(payload) > 1024 {
		payload = payload[:1024]
	}

	select {
	case s.TraceCh <- NewTraceEvent(dir, local, remote, len(payload), payload, clientID):
	default:
	}
}

func (s *Server) trace(dir TraceDirection, remote *net.UDPAddr, payload []byte) {
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
	clientID, ok := lookupClientID(remoteAddr)
	if !ok {
		clientID = 0
	}
	s.emitTrace(dir, local, remoteAddr, payload, clientID)
}
