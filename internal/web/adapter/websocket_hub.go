package adapter

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

// WebsocketMessageType defines message types
type WebsocketMessageType string

const (
	TypeLog WebsocketMessageType = "LOG"
)

// WebSocketMessage represents a websocket message
type WebSocketMessage struct {
	Type    WebsocketMessageType `json:"type"`
	Content string               `json:"content"`
}

// WebSocketHub manages websocket connections
type WebSocketHub struct {
	conns        map[*websocket.Conn]bool
	mu           sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	onAckRP      func()
	breakRPLoop  bool
	breakRPMu    sync.Mutex
}

// NewWebSocketHub creates a new websocket hub
func NewWebSocketHub(ctx context.Context) *WebSocketHub {
	hubCtx, cancel := context.WithCancel(ctx)
	return &WebSocketHub{
		conns:  make(map[*websocket.Conn]bool),
		mu:     sync.Mutex{},
		ctx:    hubCtx,
		cancel: cancel,
	}
}

// SetOnAckRP sets the callback for ack/rp messages
func (wh *WebSocketHub) SetOnAckRP(fn func()) {
	wh.onAckRP = fn
}

// HandleWS handles a new websocket connection
func (wh *WebSocketHub) HandleWS(ws *websocket.Conn) {
	wh.mu.Lock()
	wh.conns[ws] = true
	wh.mu.Unlock()

	wh.wg.Add(1)
	defer func() {
		wh.mu.Lock()
		delete(wh.conns, ws)
		wh.mu.Unlock()
		ws.Close()
		wh.wg.Done()
	}()

	wh.readLoop(ws)
}

func (wh *WebSocketHub) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		select {
		case <-wh.ctx.Done():
			return
		default:
		}

		err := ws.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		if err != nil {
			log.WithField("caller", "adapter").WithError(err).Error("readLoop error")
		}

		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.WithField("caller", "adapter").WithError(err).Error("readLoop error")
			break
		}

		msg := buf[:n]
		if string(msg) == "ack/rp" {
			wh.breakRPMu.Lock()
			wh.breakRPLoop = true
			wh.breakRPMu.Unlock()
			if wh.onAckRP != nil {
				wh.onAckRP()
			}
		}
		wh.Broadcast(msg)
	}
}

// Broadcast sends a message to all connected clients
func (wh *WebSocketHub) Broadcast(b []byte) {
	wh.mu.Lock()
	conns := make([]*websocket.Conn, 0, len(wh.conns))
	for ws := range wh.conns {
		conns = append(conns, ws)
	}
	wh.mu.Unlock()

	for _, ws := range conns {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				log.WithField("caller", "adapter").WithError(err).Error("broadcast error")
			}
		}(ws)
	}
}

// Cancel cancels the hub context
func (wh *WebSocketHub) Cancel() {
	wh.cancel()
}

// Wait waits for all goroutines to finish
func (wh *WebSocketHub) Wait() {
	wh.wg.Wait()
}

// ShouldBreakRPLoop returns whether the RP loop should break
func (wh *WebSocketHub) ShouldBreakRPLoop() bool {
	wh.breakRPMu.Lock()
	defer wh.breakRPMu.Unlock()
	return wh.breakRPLoop
}
