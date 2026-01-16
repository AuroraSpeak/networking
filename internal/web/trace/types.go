package trace

import "time"

type Direction string

const (
	In  Direction = "in"
	Out Direction = "out"
)

// Event represents a single trace event for network communication
type Event struct {
	TS       time.Time `json:"ts"`
	Dir      Direction `json:"dir"`
	Local    string    `json:"local"`
	Remote   string    `json:"remote"`
	Len      int       `json:"len"`
	Payload  []byte    `json:"payload"`
	ClientID int       `json:"client_id"`
}

// NewEvent creates a new trace event with current timestamp
func NewEvent(dir Direction, local, remote string, length int, payload []byte, clientID int) Event {
	return Event{
		TS:       time.Now(),
		Dir:      dir,
		Local:    local,
		Remote:   remote,
		Len:      length,
		Payload:  payload,
		ClientID: clientID,
	}
}
