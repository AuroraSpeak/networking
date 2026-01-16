package adapter

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/aura-speak/networking/internal/web/orchestrator"
	"github.com/aura-speak/networking/internal/web/sniffer"
)

// SnifferHandlers contains HTTP handlers for packet sniffer operations
type SnifferHandlers struct {
	snifferService *sniffer.Service
	clientService  *orchestrator.UDPClientService
}

// NewSnifferHandlers creates new sniffer handlers
func NewSnifferHandlers(snifferSvc *sniffer.Service, clientSvc *orchestrator.UDPClientService) *SnifferHandlers {
	return &SnifferHandlers{
		snifferService: snifferSvc,
		clientService:  clientSvc,
	}
}

// SnifferPacketResponse represents a packet in the API response
type SnifferPacketResponse struct {
	TS         string `json:"ts"`
	Dir        string `json:"dir"`
	Local      string `json:"local"`
	Remote     string `json:"remote"`
	Payload    string `json:"payload"` // Base64 encoded
	ClientID   int    `json:"client_id"`
	PacketType string `json:"packet_type,omitempty"`
}

// PacketsResponse represents the response for GetPackets
type PacketsResponse struct {
	Packets []SnifferPacketResponse `json:"packets"`
}

// GetPackets returns all captured packets, optionally filtered by client name
func (h *SnifferHandlers) GetPackets(w http.ResponseWriter, r *http.Request) {
	var packets []sniffer.Packet

	clientName := r.URL.Query().Get("name")
	if clientName != "" {
		uc, ok := h.clientService.GetClient(clientName)
		if !ok {
			apiError := ApiError{
				Code:    http.StatusNotFound,
				Message: "Client not found",
			}
			apiError.Send(w)
			return
		}
		packets = h.snifferService.GetByClientID(uc.ID)
	} else {
		// Return all packets including those without ClientID
		packets = h.snifferService.GetAll()
	}

	// Convert to response format with Base64 encoded payload
	response := PacketsResponse{
		Packets: make([]SnifferPacketResponse, len(packets)),
	}

	for i, p := range packets {
		response.Packets[i] = SnifferPacketResponse{
			TS:         p.TS.Format("2006-01-02T15:04:05.000Z07:00"),
			Dir:        string(p.Dir),
			Local:      p.Local,
			Remote:     p.Remote,
			Payload:    base64.StdEncoding.EncodeToString(p.Payload),
			ClientID:   p.ClientID,
			PacketType: p.PacketType,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		apiError := ApiError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to encode response",
			Details: err.Error(),
		}
		apiError.Send(w)
		return
	}
}

// ClearPackets removes all captured packets
func (h *SnifferHandlers) ClearPackets(w http.ResponseWriter, r *http.Request) {
	h.snifferService.Clear()
	apiSuccess := ApiSuccess{
		Message: "All packets cleared",
	}
	apiSuccess.Send(w)
}
