package web

import (
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
	apiSuccess := ApiSuccess{
		Message: "UDP server started",
	}
	apiSuccess.Send(w)
}

// StopUDPServer stops the running UDP server
func (s *Server) stopUDPServer(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	udpServer := s.udpServer
	s.mu.Unlock()

	if udpServer == nil {
		log.Warn("UDP server is not running")
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "UDP server is not running",
		}
		apiError.Send(w)
		return
	}

	udpServer.Stop()
	log.Info("UDP server stopped")
	apiSuccess := ApiSuccess{
		Message: "UDP server stopped",
	}
	apiSuccess.Send(w)
}

func (s *Server) getUDPServerState(w http.ResponseWriter, r *http.Request) {
	if s.udpServer == nil {
		serverStateResponse := ServerStateResponse{
			ShouldStop: false,
			IsAlive:    false,
		}
		serverStateResponse.Send(w)
		return
	}
	s.mu.Lock()
	state := s.udpServer.ServerState
	s.mu.Unlock()

	serverStateResponse := ServerStateResponse{
		ShouldStop: state.ShouldStop,
		IsAlive:    state.IsAlive,
	}
	serverStateResponse.Send(w)
}
