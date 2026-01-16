package adapter

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/aura-speak/networking/internal/web/orchestrator"
	log "github.com/sirupsen/logrus"
)

// ClientHandlers contains HTTP handlers for UDP client operations
type ClientHandlers struct {
	clientService *orchestrator.UDPClientService
	wsHub         *WebSocketHub
	onClientStart func(name string, id int) error
}

// NewClientHandlers creates new client handlers
func NewClientHandlers(svc *orchestrator.UDPClientService, hub *WebSocketHub, onStart func(name string, id int) error) *ClientHandlers {
	return &ClientHandlers{
		clientService: svc,
		wsHub:         hub,
		onClientStart: onStart,
	}
}

// StartUDPClient creates and starts a new UDP client
func (h *ClientHandlers) StartUDPClient(w http.ResponseWriter, r *http.Request) {
	name, id, _, err := h.clientService.CreateClient(func(name string, payload []byte) error {
		return h.handleClientPacket(name, payload)
	})
	if err != nil {
		apiError := ApiError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create UDP client",
			Details: err.Error(),
		}
		apiError.Send(w)
		return
	}

	if h.onClientStart != nil {
		if err := h.onClientStart(name, id); err != nil {
			log.WithField("caller", "adapter").WithError(err).Error("onClientStart callback failed")
		}
	}

	h.wsHub.Broadcast([]byte("cnu"))
	udpClientResponse := UDPClientResponse{
		Name: name,
		Id:   id,
	}
	udpClientResponse.Send(w)
}

func (h *ClientHandlers) handleClientPacket(name string, payload []byte) error {
	if err := h.clientService.AddDatagram(name, orchestrator.ServerToClient, payload); err != nil {
		return err
	}

	uc, ok := h.clientService.GetClient(name)
	if !ok {
		return nil
	}

	if h.wsHub != nil {
		h.wsHub.Broadcast([]byte("usu" + strconv.Itoa(uc.ID)))
	}
	return nil
}

// StopUDPClient stops a UDP client
func (h *ClientHandlers) StopUDPClient(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "Name is required",
		}
		apiError.Send(w)
		return
	}

	if err := h.clientService.StopClient(name); err != nil {
		apiError := ApiError{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}
		apiError.Send(w)
		return
	}

	apiSuccess := ApiSuccess{
		Message: "UDP client stopped",
	}
	apiSuccess.Send(w)
}

// GetUDPClientStateByName returns a client's state by name
func (h *ClientHandlers) GetUDPClientStateByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "Name is required",
		}
		apiError.Send(w)
		return
	}

	uc, ok := h.clientService.GetClient(name)
	if !ok {
		apiError := ApiError{
			Code:    http.StatusNotFound,
			Message: "UDP client not found",
		}
		apiError.Send(w)
		return
	}

	udpClientStateResponse := UDPClientStateResponse{
		Id:        uc.ID,
		Running:   uc.Running,
		Datagrams: uc.Datagrams,
	}
	udpClientStateResponse.Send(w)
}

// GetUDPClientStateById returns a client's state by ID
func (h *ClientHandlers) GetUDPClientStateById(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "ID is required",
		}
		apiError.Send(w)
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "ID is invalid",
		}
		apiError.Send(w)
		return
	}

	uc, _, ok := h.clientService.GetClientByID(idInt)
	if !ok {
		apiError := ApiError{
			Code:    http.StatusNotFound,
			Message: "UDP client not found",
		}
		apiError.Send(w)
		return
	}

	udpClientStateResponse := UDPClientStateResponse{
		Id:        uc.ID,
		Running:   uc.Running,
		Datagrams: uc.Datagrams,
	}
	udpClientStateResponse.Send(w)
}

// GetAllUDPClients returns all UDP clients
func (h *ClientHandlers) GetAllUDPClients(w http.ResponseWriter, r *http.Request) {
	clients := h.clientService.GetAllClients()
	allUDPClientResponse := AllUDPClientResponse{
		UDPClients: []UDPClientResponse{},
	}
	for name, uc := range clients {
		allUDPClientResponse.UDPClients = append(allUDPClientResponse.UDPClients, UDPClientResponse{
			Id:   uc.ID,
			Name: name,
		})
	}
	allUDPClientResponse.Send(w)
}

// GetAllUDPClientPaginated returns paginated UDP clients
func (h *ClientHandlers) GetAllUDPClientPaginated(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	pageInt := 1
	if pageStr != "" {
		var err error
		pageInt, err = strconv.Atoi(pageStr)
		if err != nil || pageInt < 1 {
			apiError := ApiError{
				Code:    http.StatusBadRequest,
				Message: "Page must be a positive integer",
			}
			apiError.Send(w)
			return
		}
	}

	pageSizeStr := r.URL.Query().Get("pageSize")
	pageSizeInt := 10
	if pageSizeStr != "" {
		var err error
		pageSizeInt, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSizeInt < 1 {
			apiError := ApiError{
				Code:    http.StatusBadRequest,
				Message: "Page size must be a positive integer",
			}
			apiError.Send(w)
			return
		}
	}

	searchQuery := r.URL.Query().Get("q")

	clients := h.clientService.GetAllClients()
	allItems := []UDPClientListItem{}
	for name, uc := range clients {
		if searchQuery != "" {
			if !strings.Contains(strings.ToLower(name), strings.ToLower(searchQuery)) {
				continue
			}
		}
		allItems = append(allItems, UDPClientListItem{
			Id:   uc.ID,
			Name: name,
		})
	}

	total := len(allItems)
	totalPages := int(math.Ceil(float64(total) / float64(pageSizeInt)))

	if pageInt > totalPages && totalPages > 0 {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "Page is out of range",
		}
		apiError.Send(w)
		return
	}

	startIndex := (pageInt - 1) * pageSizeInt
	endIndex := startIndex + pageSizeInt
	if endIndex > total {
		endIndex = total
	}

	var items []UDPClientListItem
	if startIndex < total {
		items = allItems[startIndex:endIndex]
	} else {
		items = []UDPClientListItem{}
	}

	paginatedResponse := UDPClientPaginatedResponse{
		Items:    items,
		Page:     pageInt,
		PageSize: pageSizeInt,
		Total:    total,
	}
	paginatedResponse.Send(w)
}

// SendDatagram sends a datagram through a client
func (h *ClientHandlers) SendDatagram(w http.ResponseWriter, r *http.Request) {
	var req SendDatagramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Details: err.Error(),
		}
		apiError.Send(w)
		return
	}

	if req.Format != "hex" && req.Format != "text" {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "Format must be 'hex' or 'text'",
		}
		apiError.Send(w)
		return
	}

	if err := h.clientService.SendDatagram(req.Id, req.Message, req.Format); err != nil {
		apiError := ApiError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
		apiError.Send(w)
		return
	}

	if h.wsHub != nil {
		h.wsHub.Broadcast([]byte("usu" + strconv.Itoa(req.Id)))
	}

	response := SendDatagramResponse{
		Message: "Datagram sent successfully",
	}
	log.Infof("Datagram sent successfully")
	response.Send(w)
}
