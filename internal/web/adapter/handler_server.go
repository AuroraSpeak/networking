package adapter

import (
	"net/http"

	"github.com/aura-speak/networking/internal/web/orchestrator"
	log "github.com/sirupsen/logrus"
)

// ServerHandlers contains HTTP handlers for UDP server operations
type ServerHandlers struct {
	serverService *orchestrator.UDPServerService
	wsHub         *WebSocketHub
	onServerStart func() error
}

// NewServerHandlers creates new server handlers
func NewServerHandlers(svc *orchestrator.UDPServerService, hub *WebSocketHub, onStart func() error) *ServerHandlers {
	return &ServerHandlers{
		serverService: svc,
		wsHub:         hub,
		onServerStart: onStart,
	}
}

// StartUDPServer starts the UDP server
func (h *ServerHandlers) StartUDPServer(w http.ResponseWriter, r *http.Request) {
	if h.onServerStart != nil {
		if err := h.onServerStart(); err != nil {
			apiError := ApiError{
				Code:    http.StatusInternalServerError,
				Message: "Failed to start UDP server",
				Details: err.Error(),
			}
			apiError.Send(w)
			return
		}
	}

	apiSuccess := ApiSuccess{
		Message: "UDP server started",
	}
	apiSuccess.Send(w)
}

// StopUDPServer stops the UDP server
func (h *ServerHandlers) StopUDPServer(w http.ResponseWriter, r *http.Request) {
	if !h.serverService.IsRunning() {
		log.WithField("caller", "adapter").Warn("UDP server is not running")
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "UDP server is not running",
		}
		apiError.Send(w)
		return
	}

	h.serverService.Stop()
	log.WithField("caller", "adapter").Info("UDP server stopped")
	apiSuccess := ApiSuccess{
		Message: "UDP server stopped",
	}
	apiSuccess.Send(w)
}

// GetUDPServerState returns the current server state
func (h *ServerHandlers) GetUDPServerState(w http.ResponseWriter, r *http.Request) {
	shouldStop, isAlive := h.serverService.GetState()
	serverStateResponse := ServerStateResponse{
		ShouldStop: shouldStop,
		IsAlive:    isAlive,
	}
	serverStateResponse.Send(w)
}
