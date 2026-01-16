package adapter

import (
	"encoding/json"
	"net/http"

	"github.com/aura-speak/networking/internal/web/orchestrator"
	log "github.com/sirupsen/logrus"
)

// ApiResponse interface for all API responses
type ApiResponse interface {
	Send(w http.ResponseWriter)
}

// ApiError represents an error response
type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}

func (e *ApiError) Send(w http.ResponseWriter) {
	w.WriteHeader(e.Code)
	b, err := json.Marshal(e)
	if err != nil {
		log.WithField("caller", "adapter").WithError(err).Error("Can't marshal ApiError to json")
	}
	w.Write(b)
	w.Write([]byte("\n"))
}

// ApiSuccess represents a success response
type ApiSuccess struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (s *ApiSuccess) Send(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(s)
	if err != nil {
		log.WithField("caller", "adapter").WithError(err).Error("Can't marshal ApiSuccess to json")
	}
	w.Write(b)
	w.Write([]byte("\n"))
}

// ServerStateResponse represents UDP server state
type ServerStateResponse struct {
	ShouldStop bool `json:"shouldStop"`
	IsAlive    bool `json:"isAlive"`
}

func (s *ServerStateResponse) Send(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(s)
	if err != nil {
		log.WithField("caller", "adapter").WithError(err).Error("Can't marshal ServerStateResponse to json")
	}
	w.Write(b)
	w.Write([]byte("\n"))
}

// UDPClientResponse represents a UDP client
type UDPClientResponse struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

func (s *UDPClientResponse) Send(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(s)
	if err != nil {
		log.WithField("caller", "adapter").WithError(err).Error("Can't marshal UDPClientResponse to json")
	}
	w.Write(b)
	w.Write([]byte("\n"))
}

// UDPClientStateResponse represents a UDP client's state
type UDPClientStateResponse struct {
	Id        int                     `json:"id"`
	Running   bool                    `json:"running"`
	Datagrams []orchestrator.Datagram `json:"datagrams"`
}

func (s *UDPClientStateResponse) Send(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(s)
	if err != nil {
		log.WithField("caller", "adapter").WithError(err).Error("Can't marshal UDPClientStateResponse to json")
	}
	w.Write(b)
	w.Write([]byte("\n"))
}

// AllUDPClientResponse represents all UDP clients
type AllUDPClientResponse struct {
	UDPClients []UDPClientResponse `json:"udpClients"`
}

func (a *AllUDPClientResponse) Send(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(a)
	if err != nil {
		log.WithField("caller", "adapter").WithError(err).Error("Can't marshal AllUDPClientResponse to json")
	}
	w.Write(b)
	w.Write([]byte("\n"))
}

// UDPClientListItem for paginated responses
type UDPClientListItem struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// UDPClientPaginatedResponse for paginated client list
type UDPClientPaginatedResponse struct {
	Items    []UDPClientListItem `json:"items"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
	Total    int                 `json:"total"`
}

func (p *UDPClientPaginatedResponse) Send(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(p)
	if err != nil {
		log.WithField("caller", "adapter").WithError(err).Error("Can't marshal PaginatedResponse to json")
	}
	w.Write(b)
	w.Write([]byte("\n"))
}

// SendDatagramRequest for sending datagrams
type SendDatagramRequest struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
	Format  string `json:"format"` // "hex" or "text"
}

// SendDatagramResponse after sending a datagram
type SendDatagramResponse struct {
	Message string `json:"message"`
}

func (s *SendDatagramResponse) Send(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(s)
	if err != nil {
		log.WithField("caller", "adapter").WithError(err).Error("Can't marshal SendDatagramResponse to json")
	}
	w.Write(b)
	w.Write([]byte("\n"))
}

// MermaidResponse for trace diagrams
type MermaidResponse struct {
	Heading string `json:"heading"`
	Diagram string `json:"diagram"`
}

func (m *MermaidResponse) Send(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(m)
	if err != nil {
		log.WithField("caller", "adapter").WithError(err).Error("Can't marshal MermaidDiagram to json")
	}
	w.Write(b)
	w.Write([]byte("\n"))
}

// DatagramResponse for datagram details
type DatagramResponse struct {
	Direction orchestrator.DatagramDirection `json:"direction"`
	Message   []byte                         `json:"message"`
}

func (d *DatagramResponse) Send(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(d)
	if err != nil {
		log.WithField("caller", "adapter").WithError(err).Error("Can't marshal Datagram to json")
	}
	w.Write(b)
	w.Write([]byte("\n"))
}
