package web

import (
	"net/http"

	"golang.org/x/net/websocket"
)

func (s *Server) registerRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(s.handleWS))
	mux.Handle("/", http.FileServer(http.Dir("./bin")))

	// UDP Server handlers
	mux.HandleFunc("POST /api/server/start", s.startUDPServer)
	mux.HandleFunc("POST /api/server/stop", s.stopUDPServer)
	mux.HandleFunc("GET /api/server/get", s.getUDPServerState)

	mux.HandleFunc("POST /api/client/start", s.startUDPClient)
	mux.HandleFunc("POST /api/client/stop", s.stopUDPClient)
	mux.HandleFunc("POST /api/client/send", s.sendDatagram)
	mux.HandleFunc("GET /api/client/get/name", s.getUDPClientStateByName)
	mux.HandleFunc("GET /api/client/get/id", s.getUDPClientStateById)
	mux.HandleFunc("GET /api/client/get/all", s.getAllUDPClients)

	// Trace handlers
	mux.HandleFunc("GET /api/traces/all", s.getTraces)
	// Paginated all UDP clients
	mux.HandleFunc("GET /api/client/get/all/paginated", s.getAllUDPClientPaginated)

	// Wrap the entire mux with CORS for all /api/ routes
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply CORS to /api/ routes
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			corsWrapper(mux).ServeHTTP(w, r)
		} else {
			mux.ServeHTTP(w, r)
		}
	})
}
