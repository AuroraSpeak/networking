package adapter

import (
	"fmt"
	"net/http"

	"github.com/aura-speak/networking/internal/web/orchestrator"
	"github.com/aura-speak/networking/internal/web/trace"
)

// TraceHandlers contains HTTP handlers for trace operations
type TraceHandlers struct {
	traceService  *trace.Service
	clientService *orchestrator.UDPClientService
}

// NewTraceHandlers creates new trace handlers
func NewTraceHandlers(traceSvc *trace.Service, clientSvc *orchestrator.UDPClientService) *TraceHandlers {
	return &TraceHandlers{
		traceService:  traceSvc,
		clientService: clientSvc,
	}
}

// GetTraces returns traces for a specific client as a Mermaid diagram
func (h *TraceHandlers) GetTraces(w http.ResponseWriter, r *http.Request) {
	clientName := r.URL.Query().Get("name")
	if clientName == "" {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "Client name is required",
		}
		apiError.Send(w)
		return
	}

	uc, ok := h.clientService.GetClient(clientName)
	if !ok {
		apiError := ApiError{
			Code:    http.StatusNotFound,
			Message: "Client not found",
		}
		apiError.Send(w)
		return
	}

	traces := h.traceService.GetByClientID(uc.ID)
	md := trace.BuildSequenceDiagram(traces)

	traceRes := MermaidResponse{
		Heading: fmt.Sprintf("Diagram for user: %s", clientName),
		Diagram: md,
	}
	traceRes.Send(w)
}
