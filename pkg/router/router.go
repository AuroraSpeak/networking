// Package Router contains the core routing for the packets
// It is responsible for routing packets to the appropriate handlers
// The implementation is based on the UDP protocol
// It Handles Packet Routing, Starts and Runs the Router
// Add callbacks to the Router to handle incoming packets
// It may contains some default callbacks required for handling the router state
package router

import "errors"

// PacketHandler is the function type for the PacketHandler
// It takes a packet and returns an error
type PacketHandler func(msg []byte) error

// PacketRouter is the main struct for the PacketRouter
// It contains the handlers for the PacketRouter
type PacketRouter struct {
	handlers map[string]PacketHandler
}

// NewPacketRouter creates a new PacketRouter
func NewPacketRouter() *PacketRouter {
	return &PacketRouter{
		handlers: make(map[string]PacketHandler),
	}
}

// OnPacket registers a new PacketHandler for a specific packet type
//
// Example:
//
//	router.OnPacket("text", func(packet []byte) error {
//		fmt.Println("Received text packet:", string(packet))
//		return nil
//	})
func (r *PacketRouter) OnPacket(packetType string, handler PacketHandler) {
	r.handlers[packetType] = handler
}

// HandlePacket routes a packet to the appropriate handler
//
// Example:
//
// router.HandlePacket("text", []byte("Hello, Server!"))
//
// # It returns an error if no handler is found for the packet type
//
// Example:
//
// err := router.HandlePacket("text", []byte("Hello, Server!"))
//
//	if err != nil {
//		fmt.Println("Error routing packet:", err)
//	}
func (r *PacketRouter) HandlePacket(packetType string, packet []byte) error {
	handler, ok := r.handlers[packetType]
	if !ok {
		return errors.New("no handler found for packet type")
	}
	return handler(packet)
}
