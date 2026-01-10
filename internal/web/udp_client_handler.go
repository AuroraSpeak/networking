package web

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/aura-speak/networking/pkg/client"
	"github.com/aura-speak/networking/pkg/protocol"
	log "github.com/sirupsen/logrus"
)

func (s *Server) startUDPClient(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name := s.genUDPClient(s.config.UDPPort)
	s.udpClients[name].client.OnPacket(protocol.PacketTypeDebugAny, func(packet *protocol.Packet) error {
		return s.handleAllClient(name, packet.Payload)
	})
	s.shutdownWg.Go(func() {
		s.udpClients[name].client.Run()
	})
	udpClientResponse := UDPClientResponse{
		Name: name,
		Id:   s.udpClients[name].id,
	}

	s.wsHub.Broadcast([]byte("cnu"))
	udpClientResponse.Send(w)
}

func (s *Server) stopUDPClient(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name := r.URL.Query().Get("name")
	if name == "" {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "Name is required",
		}
		apiError.Send(w)
		return
	}
	udpClient, ok := s.udpClients[name]
	if !ok {
		apiError := ApiError{
			Code:    http.StatusNotFound,
			Message: "UDP client not found",
		}
		apiError.Send(w)
		return
	}
	udpClient.client.Stop()
	// Remove client command channel from map
	delete(s.clientCommandChs, udpClient.id)
	apiSuccess := ApiSuccess{
		Message: "UDP client stopped",
	}
	apiSuccess.Send(w)
}

func (s *Server) getUDPClientStateByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "Name is required",
		}
		apiError.Send(w)
		return
	}
	udpClient, ok := s.udpClients[name]
	if !ok {
		apiError := ApiError{
			Code:    http.StatusNotFound,
			Message: "UDP client not found",
		}
		apiError.Send(w)
		return
	}
	udpClientStateResponse := UDPClientStateResponse{
		Id:        udpClient.id,
		Running:   udpClient.running,
		Datagrams: udpClient.datagrams,
	}
	udpClientStateResponse.Send(w)
}

func (s *Server) getUDPClientStateById(w http.ResponseWriter, r *http.Request) {
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
	for _, udpClient := range s.udpClients {
		if udpClient.id == idInt {
			udpClientStateResponse := UDPClientStateResponse{
				Id:        udpClient.id,
				Running:   udpClient.running,
				Datagrams: udpClient.datagrams,
			}
			udpClientStateResponse.Send(w)
			return
		}
	}
	apiError := ApiError{
		Code:    http.StatusNotFound,
		Message: "UDP client not found",
	}
	apiError.Send(w)
}

func (s *Server) getAllUDPClients(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	allUDPClientResponse := AllUDPClientResponse{
		UDPClients: []UDPClientResponse{},
	}
	for name := range s.udpClients {
		allUDPClientResponse.UDPClients = append(allUDPClientResponse.UDPClients, UDPClientResponse{
			Id:   s.udpClients[name].id,
			Name: name,
		})
	}
	allUDPClientResponse.Send(w)
}

func (s *Server) getAllUDPClientPaginated(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse page parameter with default value 1
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

	// Parse pageSize parameter with default value 10
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

	// Parse search query parameter
	searchQuery := r.URL.Query().Get("q")

	// Filter and collect all matching UDP clients (only basic info: id and name)
	allItems := []UDPClientListItem{}
	for name, udpClient := range s.udpClients {
		// Apply search filter if provided
		if searchQuery != "" {
			// Case-insensitive search in client name
			if !strings.Contains(strings.ToLower(name), strings.ToLower(searchQuery)) {
				continue
			}
		}

		allItems = append(allItems, UDPClientListItem{
			Id:   udpClient.id,
			Name: name,
		})
	}

	// Calculate pagination
	total := len(allItems)
	totalPages := int(math.Ceil(float64(total) / float64(pageSizeInt)))

	// Validate page number
	if pageInt > totalPages && totalPages > 0 {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "Page is out of range",
		}
		apiError.Send(w)
		return
	}

	// Calculate slice indices for pagination
	startIndex := (pageInt - 1) * pageSizeInt
	endIndex := startIndex + pageSizeInt
	if endIndex > total {
		endIndex = total
	}

	// Extract items for current page
	var items []UDPClientListItem
	if startIndex < total {
		items = allItems[startIndex:endIndex]
	} else {
		items = []UDPClientListItem{}
	}

	// Create and send response
	paginatedResponse := UDPClientPaginatedRespone{
		Items:    items,
		Page:     pageInt,
		PageSize: pageSizeInt,
		Total:    total,
	}
	paginatedResponse.Send(w)
}

func (s *Server) sendDatagram(w http.ResponseWriter, r *http.Request) {
	// Parse request body
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

	// Validate format
	if req.Format != "hex" && req.Format != "text" {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "Format must be 'hex' or 'text'",
		}
		apiError.Send(w)
		return
	}

	// Find client by ID and validate (with lock)
	s.mu.Lock()
	var clientToSend *client.Client
	var clientName string
	var clientRunning bool
	for name, uc := range s.udpClients {
		if uc.id == req.Id {
			clientToSend = uc.client
			clientName = name
			clientRunning = uc.running
			break
		}
	}
	s.mu.Unlock()

	if clientToSend == nil {
		apiError := ApiError{
			Code:    http.StatusNotFound,
			Message: "UDP client not found",
		}
		apiError.Send(w)
		return
	}

	// Check if client is running
	if !clientRunning {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "Client is not running",
		}
		apiError.Send(w)
		return
	}

	// Convert message to []byte based on format
	messageBytes, err := convertMessageToBytes(req.Message, req.Format)
	if err != nil {
		apiError := ApiError{
			Code:    http.StatusBadRequest,
			Message: "Invalid hex string",
			Details: err.Error(),
		}
		apiError.Send(w)
		return
	}

	// Send message via client (außerhalb des Locks, damit es nicht blockiert)
	if err := clientToSend.Send(messageBytes); err != nil {
		apiError := ApiError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to send datagram",
			Details: err.Error(),
		}
		apiError.Send(w)
		return
	}

	// Lock wieder holen für Map-Update
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find client again (könnte sich geändert haben)
	udpClient, ok := s.udpClients[clientName]
	if !ok {
		apiError := ApiError{
			Code:    http.StatusNotFound,
			Message: "UDP client not found",
		}
		apiError.Send(w)
		return
	}

	// Store datagram in client's datagrams list
	newDatagram := datagram{
		Direction: ClientToServer,
		Message:   messageBytes,
	}
	udpClient.datagrams = append(udpClient.datagrams, newDatagram)
	s.udpClients[clientName] = udpClient

	// Broadcast WebSocket update
	if s.wsHub != nil {
		s.wsHub.Broadcast([]byte("usu" + strconv.Itoa(req.Id)))
	}

	// Send success response
	response := SendDatagramResponse{
		Message: "Datagram sent successfully",
	}
	log.WithField("caller", "web").Infof("Datagram sent successfully: %s", string(messageBytes))
	response.Send(w)
}
