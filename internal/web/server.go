package web

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aura-speak/networking/pkg/client"
	"github.com/aura-speak/networking/pkg/server"
	"golang.org/x/net/websocket"
)

type Config struct {
	UDPPort int
}

type Server struct {
	Port       int
	mu         sync.Mutex
	config     Config
	httpServer *http.Server
	ctx        context.Context
	cancel     context.CancelFunc
	shutdownWg sync.WaitGroup

	// WebSocket Hub
	wsHub *WebSocketHub

	// UDP Parts
	udpServer  *server.Server
	udpClients map[*client.Client]bool
	// Communicate from UDP Server and Clients to WebSocket Hub
	messageCh chan []InternalMessage
}

func NewServer(port int, udpPort int) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		Port:       port,
		mu:         sync.Mutex{},
		ctx:        ctx,
		cancel:     cancel,
		wsHub:      NewWebSocketHub(ctx),
		config:     Config{UDPPort: udpPort},
		udpClients: map[*client.Client]bool{},
	}
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(s.handleWS))
	mux.Handle("/", http.FileServer(http.Dir("./bin")))

	// UDP Server handlers
	mux.HandleFunc("/udp/server-start", s.startUDPServer)
	mux.HandleFunc("/udp/server-stop", s.stopUDPServer)
	mux.HandleFunc("/udp/server-state", s.getUDPServerState)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
	}

	s.shutdownWg.Go(func() {
		s.handleInternal()
	})

	fmt.Printf("Starting server on http://localhost:%d\n", s.Port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) handleWS(ws *websocket.Conn) {
	if s.wsHub != nil {
		s.wsHub.HandleWS(ws)
	}
}

// Handles all internal communications to the web server
// Avalible Commands:
// uss: tells the clients, that the server state has been updated
func (s *Server) handleInternal() {
	s.shutdownWg.Go(func() {
		// Brodcast through on UDP Server State Changes
		for {
			// Wait until UDP server is initialized
			s.mu.Lock()
			udpServer := s.udpServer
			s.mu.Unlock()

			if udpServer == nil {
				// UDP server not started yet, wait a bit and check again
				select {
				case <-s.ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}

			// UDP server is available, listen for commands
			select {
			case cmd := <-udpServer.OutCommandCh:
				switch cmd {
				case server.CmdUpdateServerState:
					// Broadcast to all WebSocket Clients that the UDP Server State has changed
					if s.wsHub != nil {
						s.wsHub.Broadcast([]byte("uss"))
					}
				}
			case <-s.ctx.Done():
				return
			}
		}
	})
}

func (s *Server) Shutdown(timeout time.Duration) error {
	fmt.Println("Shutting down server...")

	// Signals all go routines to cancel
	s.cancel()

	// Cancel the WebSocketHub
	if s.wsHub != nil {
		s.wsHub.Cancel()
	}

	// Wait for WebSocketHub goroutines to finish
	if s.wsHub != nil {
		done := make(chan struct{})
		go func() {
			s.wsHub.Wait()
			close(done)
		}()

		select {
		case <-done:
			fmt.Println("WebSocketHub goroutines finished")
		case <-time.After(timeout):
			fmt.Println("Warning: WebSocketHub wait timeout reached")
		}
	}

	// Waits for all Goroutines with Timeout
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
