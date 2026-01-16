package orchestrator

import "sync/atomic"

// IDCounter provides thread-safe ID generation
type IDCounter struct {
	counter int32
}

// NewIDCounter creates a new ID counter
func NewIDCounter() *IDCounter {
	return &IDCounter{}
}

// GetNextID returns the next unique ID
func (c *IDCounter) GetNextID() int {
	return int(atomic.AddInt32(&c.counter, 1))
}
