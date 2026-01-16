//go:build debug
// +build debug

package adapter

import (
	"time"

	"github.com/aura-speak/networking/internal/web/trace"
)

func (s *Server) handleTraceEvents() {
	for {
		srv := s.serverService.GetServer()
		if srv == nil {
			select {
			case <-s.ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}

		select {
		case <-s.ctx.Done():
			return
		case serverTrace := <-srv.TraceCh:
			// Convert server.TraceEvent to trace.Event
			event := trace.Event{
				TS:       serverTrace.TS,
				Dir:      trace.Direction(serverTrace.Dir),
				Local:    serverTrace.Local,
				Remote:   serverTrace.Remote,
				Len:      serverTrace.Len,
				Payload:  serverTrace.Payload,
				ClientID: serverTrace.ClientID,
			}
			s.traceService.Add(event)
		}
	}
}
