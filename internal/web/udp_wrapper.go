package web

import (
	"net/http"

	"github.com/aura-speak/networking/pkg/server"

	log "github.com/sirupsen/logrus"
)

// StartUDPServer starts a UDP server in its own goroutine that listens for incoming messages
func (s *Server) startUDPServer(w http.ResponseWriter, r *http.Request) {
	s.udpServer = server.NewServer(s.config.UDPPort, s.ctx)
	var err error
	s.shutdownWg.Go(func() {
		defer s.shutdownWg.Done()
		if err = s.udpServer.Run(); err != nil {
			log.WithError(err).Error("error starting udp server")
		}
	})

	log.Infof("UDP server started on port %d", s.config.UDPPort)
	w.WriteHeader(http.StatusOK)
}

// StopUDPServer stops the running UDP server
func (s *Server) stopUDPServer(w http.ResponseWriter, r *http.Request) {
	if s.udpServer == nil {
		log.Warn("UDP server is not running")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.udpServer.Stop()
	log.Info("UDP server stopped")
	w.WriteHeader(http.StatusOK)
}
