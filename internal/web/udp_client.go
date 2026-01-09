package web

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	webutil "github.com/aura-speak/networking/internal/web/utils"
	"github.com/aura-speak/networking/pkg/client"
	log "github.com/sirupsen/logrus"
)

type ids struct {
	nextID int
	mu     sync.Mutex
}

var idCounter ids

type datagramDirection int

const (
	ClientToServer = 1
	ServerToClient = 2
)

type datagram struct {
	Direction datagramDirection
	message   []byte
}

func init() {
	idCounter = ids{
		nextID: 0,
		mu:     sync.Mutex{},
	}
}

func (i *ids) getNextID() int {
	i.mu.Lock()
	id := i.nextID
	i.nextID++
	i.mu.Unlock()
	return id
}

type udpClient struct {
	// ID of the client
	id int
	// The UDP client
	client *client.Client
	// name random chosen
	name string
	// Datagram by the user
	datagrams []datagram
	// is it running
	running bool
}

func (s *Server) genUDPClient(port int) string {
	name := webutil.GetFirstName()
	id := idCounter.getNextID()
	client := client.NewDebugClient("localhost", port, id)
	s.udpClients[name] = udpClient{
		id:        id,
		client:    client,
		name:      name,
		datagrams: []datagram{},
	}
	// Register client command channel and start listening
	s.clientCommandChs[id] = client.OutCommandCh
	s.handleClientCommands(id, client.OutCommandCh)
	log.Infof("UDP client started: %s with id %d", name, id)
	return name
}

func (s *Server) startUDPClient(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name := s.genUDPClient(s.Port)
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
