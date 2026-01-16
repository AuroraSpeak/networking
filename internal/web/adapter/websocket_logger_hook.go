package adapter

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// WebSocketHook forwards log entries to websocket clients
type WebSocketHook struct {
	hub       *WebSocketHub
	formatter logrus.Formatter
}

// NewWebSocketHook creates a new websocket logger hook
func NewWebSocketHook(hub *WebSocketHub) *WebSocketHook {
	return &WebSocketHook{
		hub:       hub,
		formatter: &logrus.JSONFormatter{},
	}
}

// Levels returns all log levels
func (wh *WebSocketHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire sends the log entry to websocket clients
func (wh *WebSocketHook) Fire(entry *logrus.Entry) error {
	if wh.hub == nil {
		return nil
	}
	b, err := wh.formatter.Format(entry)
	if err != nil {
		fmt.Println(err)
		return err
	}
	wh.hub.Broadcast([]byte(b))
	return nil
}
