package web

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/aura-speak/networking/pkg/server"
	"golang.org/x/net/websocket"
)

type Server struct {
	Port  int
	conns map[*websocket.Conn]bool
	mu    sync.Mutex

	httpServer *http.Server
	ctx        context.Context
	cancel     context.CancelFunc
	shutdownWg sync.WaitGroup

	// UDP Parts
	udpServer *server.Server
}

func NewServer(port int) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		Port:   port,
		conns:  make(map[*websocket.Conn]bool),
		mu:     sync.Mutex{},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	s.shutdownWg.Add(1)
	defer s.shutdownWg.Done()

	s.mu.Lock()
	s.conns[ws] = true
	s.mu.Unlock()

	s.readLoop(ws)

	s.mu.Lock()
	delete(s.conns, ws)
	s.mu.Unlock()
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// Setze ReadDeadline, damit wir regelmäßig den Context prüfen können
		ws.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			// Timeout-Fehler ignorieren, wir prüfen dann den Context
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			fmt.Println("read error", err)
			break
		}
		msg := buf[:n]
		fmt.Println(string(msg))
		s.brodcast(msg)
	}
}

func (s *Server) brodcast(b []byte) {
	s.mu.Lock()
	conns := make([]*websocket.Conn, 0, len(s.conns))
	for ws := range s.conns {
		conns = append(conns, ws)
	}
	s.mu.Unlock()

	for _, ws := range conns {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				fmt.Println("broadcast error:", err)
			}
		}(ws)
	}
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(s.handleWS))

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
	}

	fmt.Println("Starting server on port", s.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(timeout time.Duration) error {
	fmt.Println("Shutting down server...")

	// Signal allen Goroutines, dass sie stoppen sollen
	s.cancel()

	// Schließe alle WebSocket-Verbindungen
	s.mu.Lock()
	for ws := range s.conns {
		ws.Close()
	}
	s.mu.Unlock()

	// Warte auf alle Goroutines mit Timeout
	done := make(chan struct{})
	go func() {
		s.shutdownWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("All connections closed")
	case <-time.After(timeout):
		fmt.Println("Warning: Shutdown timeout reached, some connections may not have closed gracefully")
	}

	// Stoppe den HTTP-Server
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down HTTP server: %w", err)
	}

	fmt.Println("Server shutdown complete")
	return nil
}
