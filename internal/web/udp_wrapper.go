package web

import (
	"encoding/json"
	"net/http"

	"github.com/aura-speak/networking/pkg/server"

	log "github.com/sirupsen/logrus"
)

// StartUDPServer starts a UDP server in its own goroutine that listens for incoming messages
func (s *Server) startUDPServer(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.udpServer = server.NewServer(s.config.UDPPort, s.ctx)
	udpServer := s.udpServer
	s.mu.Unlock()

	var err error
	s.shutdownWg.Go(func() {
		if err = udpServer.Run(); err != nil {
			log.WithError(err).Error("error starting udp server")
		}
	})

	log.Infof("UDP server started on port %d", s.config.UDPPort)
	w.WriteHeader(http.StatusOK)
}

// StopUDPServer stops the running UDP server
func (s *Server) stopUDPServer(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	udpServer := s.udpServer
	s.mu.Unlock()

	if udpServer == nil {
		log.Warn("UDP server is not running")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	udpServer.Stop()
	log.Info("UDP server stopped")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getUDPServerState(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	udpServer := s.udpServer
	s.mu.Unlock()

	if udpServer == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"shouldStop":false,"isAlive":false}`))
		return
	}
	b, err := json.Marshal(udpServer.ServerState)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.WithError(err).Error("Can't marshal UDP Server State to json")
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
