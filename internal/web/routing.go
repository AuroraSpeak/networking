package web

import (
	"net/http"

	"golang.org/x/net/websocket"
)

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next(w, r)
	}
}

func corsWrapper(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		handler.ServeHTTP(w, r)
	})
}

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
	mux.HandleFunc("GET /api/client/get/name", s.getUDPClientStateByName)
	mux.HandleFunc("GET /api/client/get/id", s.getUDPClientStateById)
	mux.HandleFunc("GET /api/client/get/all", s.getAllUDPClients)
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
