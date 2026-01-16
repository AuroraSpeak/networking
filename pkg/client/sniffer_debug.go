//go:build debug
// +build debug

package client

// ClientSnifferCallback is a function type for capturing packets from clients
type ClientSnifferCallback func(dir ClientSnifferDirection, local, remote string, payload []byte, clientID int, packetType string)

type ClientSnifferDirection string

const (
	ClientSnifferIn  ClientSnifferDirection = "in"
	ClientSnifferOut ClientSnifferDirection = "out"
)

var clientSnifferCallback ClientSnifferCallback

// SetClientSnifferCallback sets the callback function for client packet sniffing
func SetClientSnifferCallback(cb ClientSnifferCallback) {
	clientSnifferCallback = cb
}

// capturePacket captures a packet if the callback is set
func (c *Client) capturePacket(dir ClientSnifferDirection, payload []byte, clientID int, packetType string) {
	if clientSnifferCallback == nil {
		return
	}

	local := ""
	remote := ""
	if c.conn != nil {
		if c.conn.LocalAddr() != nil {
			local = c.conn.LocalAddr().String()
		}
		if c.conn.RemoteAddr() != nil {
			remote = c.conn.RemoteAddr().String()
		}
	}
	if local == "" {
		local = "unknown"
	}
	if remote == "" {
		remote = "unknown"
	}

	clientSnifferCallback(dir, local, remote, payload, clientID, packetType)
}
