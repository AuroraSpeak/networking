package adapter

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aura-speak/networking/internal/web/orchestrator"
	"github.com/aura-speak/networking/internal/web/trace"
	"github.com/aura-speak/networking/pkg/client"
	"github.com/aura-speak/networking/pkg/protocol"
	"github.com/aura-speak/networking/pkg/server"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

// Ensure trace.Service is used (for release builds where handleTraceEvents is no-op)
var _ = (*trace.Service)(nil)

// Config holds server configuration
type Config struct {
	UDPPort int
}

// Server is the main web server
type Server struct {
	Port       int
	config     Config
	httpServer *http.Server
	ctx        context.Context
	cancel     context.CancelFunc
	shutdownWg sync.WaitGroup

	// Components
	wsHub         *WebSocketHub
	serverService *orchestrator.UDPServerService
	clientService *orchestrator.UDPClientService
	traceService  *trace.Service

	// Handlers
	serverHandlers *ServerHandlers
	clientHandlers *ClientHandlers
	traceHandlers  *TraceHandlers

	// Client factory for debug/release builds
	clientFactory orchestrator.ClientFactory
}

// NewServer creates a new web server
func NewServer(port int, udpPort int, clientFactory orchestrator.ClientFactory) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	wsHub := NewWebSocketHub(ctx)
	log.AddHook(NewWebSocketHook(wsHub))

	traceService := trace.NewService(ctx)
	serverService := orchestrator.NewUDPServerService(ctx, udpPort)
	clientService := orchestrator.NewUDPClientService(ctx, udpPort, clientFactory)

	s := &Server{
		Port:          port,
		config:        Config{UDPPort: udpPort},
		ctx:           ctx,
		cancel:        cancel,
		wsHub:         wsHub,
		serverService: serverService,
		clientService: clientService,
		traceService:  traceService,
		clientFactory: clientFactory,
	}

	// Initialize handlers
	s.serverHandlers = NewServerHandlers(serverService, wsHub, s.startUDPServerInternal)
	s.clientHandlers = NewClientHandlers(clientService, wsHub, s.startClientInternal)
	s.traceHandlers = NewTraceHandlers(traceService, clientService)

	return s
}

// Run starts the server
func (s *Server) Run() error {
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: s.registerRoutes(),
	}

	s.shutdownWg.Add(1)
	go func() {
		defer s.shutdownWg.Done()
		s.handleServerCommands()
	}()

	s.shutdownWg.Add(1)
	go func() {
		defer s.shutdownWg.Done()
		s.handleTraceEvents()
	}()

	fmt.Printf("Starting server on http://localhost:%d\n", s.Port)

	// RP loop for client keepalive
	go s.runRPLoop()

	return s.httpServer.ListenAndServe()
}

func (s *Server) runRPLoop() {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			time.Sleep(1 * time.Second)
		}
		if s.wsHub.ShouldBreakRPLoop() {
			return
		}
		s.wsHub.Broadcast([]byte("rp"))
	}
}

func (s *Server) registerRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(s.handleWS))
	mux.Handle("/", http.FileServer(http.Dir("./bin")))

	// UDP Server handlers
	mux.HandleFunc("POST /api/server/start", s.serverHandlers.StartUDPServer)
	mux.HandleFunc("POST /api/server/stop", s.serverHandlers.StopUDPServer)
	mux.HandleFunc("GET /api/server/get", s.serverHandlers.GetUDPServerState)

	// UDP Client handlers
	mux.HandleFunc("POST /api/client/start", s.clientHandlers.StartUDPClient)
	mux.HandleFunc("POST /api/client/stop", s.clientHandlers.StopUDPClient)
	mux.HandleFunc("POST /api/client/send", s.clientHandlers.SendDatagram)
	mux.HandleFunc("GET /api/client/get/name", s.clientHandlers.GetUDPClientStateByName)
	mux.HandleFunc("GET /api/client/get/id", s.clientHandlers.GetUDPClientStateById)
	mux.HandleFunc("GET /api/client/get/all", s.clientHandlers.GetAllUDPClients)
	mux.HandleFunc("GET /api/client/get/all/paginated", s.clientHandlers.GetAllUDPClientPaginated)

	// Trace handlers
	mux.HandleFunc("GET /api/traces/all", s.traceHandlers.GetTraces)

	// CORS wrapper for API routes
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			corsWrapper(mux).ServeHTTP(w, r)
		} else {
			mux.ServeHTTP(w, r)
		}
	})
}

func (s *Server) handleWS(ws *websocket.Conn) {
	if s.wsHub != nil {
		s.wsHub.HandleWS(ws)
	}
}

func (s *Server) startUDPServerInternal() error {
	srv, err := s.serverService.Start(s.handleServerPacket)
	if err != nil {
		return err
	}

	s.shutdownWg.Add(1)
	go func() {
		defer s.shutdownWg.Done()
		if err := srv.Run(); err != nil {
			log.WithField("caller", "adapter").WithError(err).Error("error running UDP server")
		}
	}()

	return nil
}

func (s *Server) handleServerPacket(clientAddr string, payload []byte) error {
	log.WithField("caller", "adapter").Infof("Received packet: %s", string(payload))

	srv := s.serverService.GetServer()
	if srv != nil {
		srv.Broadcast(&protocol.Packet{
			PacketHeader: protocol.Header{PacketType: protocol.PacketTypeDebugAny},
			Payload:      payload,
		})
	}

	if s.wsHub != nil {
		s.wsHub.Broadcast(payload)
	}
	return nil
}

func (s *Server) startClientInternal(name string, id int) error {
	uc, ok := s.clientService.GetClient(name)
	if !ok {
		return fmt.Errorf("client not found: %s", name)
	}

	s.shutdownWg.Add(1)
	go func() {
		defer s.shutdownWg.Done()
		uc.Client.Run()
	}()

	// Start listening for client commands
	if cmdCh, ok := s.clientService.GetCommandChannel(id); ok {
		s.shutdownWg.Add(1)
		go func() {
			defer s.shutdownWg.Done()
			s.handleClientCommands(id, cmdCh)
		}()
	}

	return nil
}

func (s *Server) handleClientCommands(clientID int, cmdCh chan client.InternalCommand) {
	for {
		select {
		case cmd := <-cmdCh:
			switch cmd {
			case client.CmdUpdateClientState:
				uc, _, ok := s.clientService.GetClientByID(clientID)
				if ok {
					running := uc.Client.ClientState.Running == 1
					s.clientService.UpdateClientRunning(clientID, running)
					if s.wsHub != nil {
						s.wsHub.Broadcast([]byte("usu" + fmt.Sprintf("%d", clientID)))
					}
				}
			}
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Server) handleServerCommands() {
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
		case cmd := <-srv.OutCommandCh:
			switch cmd {
			case server.CmdUpdateServerState:
				if s.wsHub != nil {
					s.wsHub.Broadcast([]byte("uss"))
				}
			}
		case <-s.ctx.Done():
			return
		}
	}
}


// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(timeout time.Duration) error {
	fmt.Println("Shutting down server...")

	s.cancel()

	if s.wsHub != nil {
		s.wsHub.Cancel()
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down HTTP server: %w", err)
	}

	fmt.Println("Server shutdown complete")
	return nil
}
