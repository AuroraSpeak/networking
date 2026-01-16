//go:build !debug
// +build !debug

package adapter

// handleTraceEvents is a no-op in release builds
func (s *Server) handleTraceEvents() {
	// Tracing is disabled in release builds
	<-s.ctx.Done()
}
