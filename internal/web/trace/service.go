package trace

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Service manages trace events collection and storage
type Service struct {
	traces  []Event
	traceMu sync.Mutex
	ctx     context.Context
}

// NewService creates a new trace service
func NewService(ctx context.Context) *Service {
	return &Service{
		traces: make([]Event, 0),
		ctx:    ctx,
	}
}

// Add appends a trace event to the collection
func (s *Service) Add(event Event) {
	s.traceMu.Lock()
	defer s.traceMu.Unlock()
	s.traces = append(s.traces, event)
	log.WithFields(log.Fields{
		"caller": "trace",
		"cid":    event.ClientID,
	}).Debugf("Received trace: %+v", event)
}

// GetByClientID returns all traces for a specific client
func (s *Service) GetByClientID(clientID int) []Event {
	s.traceMu.Lock()
	defer s.traceMu.Unlock()

	filtered := make([]Event, 0)
	for _, t := range s.traces {
		if t.ClientID == clientID {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// GetAll returns all collected traces
func (s *Service) GetAll() []Event {
	s.traceMu.Lock()
	defer s.traceMu.Unlock()

	result := make([]Event, len(s.traces))
	copy(result, s.traces)
	return result
}

// Clear removes all traces
func (s *Service) Clear() {
	s.traceMu.Lock()
	defer s.traceMu.Unlock()
	s.traces = make([]Event, 0)
}
