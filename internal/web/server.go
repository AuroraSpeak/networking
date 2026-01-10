package web

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/aura-speak/networking/pkg/client"
	"github.com/aura-speak/networking/pkg/server"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type Config struct {
	UDPPort int
}

var breakRPLoop = false

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
	udpServer *server.Server
	// udpClientWrapper
	udpClients      map[string]udpClient
	clientMu        sync.Mutex
	udpClientAction UDPClientActionData
	// Communicate from UDP Server and Clients to WebSocket Hub
	messageCh chan []InternalMessage
	// Client command channels mapped by client ID
	clientCommandChs map[int]chan client.InternalCommand

	// Traces
	traces  []server.TraceEvent
	traceMu sync.Mutex
}

func NewServer(port int, udpPort int) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		Port:             port,
		mu:               sync.Mutex{},
		ctx:              ctx,
		cancel:           cancel,
		wsHub:            NewWebSocketHub(ctx),
		config:           Config{UDPPort: udpPort},
		udpClients:       make(map[string]udpClient),
		clientCommandChs: make(map[int]chan client.InternalCommand),
		traceMu:          sync.Mutex{},
	}
}

func (s *Server) Run() error {
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: s.registerRoutes(),
	}

	s.shutdownWg.Go(func() {
		s.handleInternal()
	})

	s.shutdownWg.Go(func() {
		s.handleTrace()
	})

	fmt.Printf("Starting server on http://localhost:%d\n", s.Port)
	go func() {
		for !breakRPLoop {
			select {
			case <-s.ctx.Done():
				breakRPLoop = true
				return
			default:
				time.Sleep(1 * time.Second)
			}
			if breakRPLoop {
				return
			}
			s.wsHub.Broadcast([]byte("rp"))
		}
	}()
	return s.httpServer.ListenAndServe()
}

func (s *Server) handleWS(ws *websocket.Conn) {
	if s.wsHub != nil {
		s.wsHub.HandleWS(ws)
	}
}

// handleClientCommands listens for commands from a specific UDP client
func (s *Server) handleClientCommands(clientID int, cmdCh chan client.InternalCommand) {
	s.shutdownWg.Go(func() {
		for {
			select {
			case cmd := <-cmdCh:
				switch cmd {
				case client.CmdUpdateClientState:
					s.mu.Lock()
					// Find udpClient by ID and update running field
					for name, uc := range s.udpClients {
						if uc.id == clientID {
							// Update running field from ClientState
							running := uc.client.ClientState.Running
							uc.running = running == 1
							// Update the map entry
							s.udpClients[name] = uc

							// Broadcast to all WebSocket Clients that the UDP Client State has changed
							if s.wsHub != nil {
								s.wsHub.Broadcast([]byte("usu" + strconv.Itoa(clientID)))
							}
							break
						}
					}
					s.mu.Unlock()
				}
			case <-s.ctx.Done():
				return
			}
		}
	})
}

// Handles all internal communications to the web server
// Avalible Commands:
// uss: tells the clients, that the server state has been updated
// usu: tells the clients, that a udp client state has been updated
// cnu: tells the web server, that a new udp client has been started
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

func (s *Server) handleTrace() {
	s.shutdownWg.Go(func() {
		for {
			s.mu.Lock()
			udpServer := s.udpServer
			s.mu.Unlock()

			if udpServer == nil {
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
			case trace := <-s.udpServer.TraceCh:
				s.traceMu.Lock()
				s.traces = append(s.traces, trace)
				log.WithFields(log.Fields{
					"caller": "web",
					"cid":    trace.ClientID,
				}).Debugf("Received trace: %+v", trace)
				s.traceMu.Unlock()
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
