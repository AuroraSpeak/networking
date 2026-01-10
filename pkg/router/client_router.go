// Package Router contains the core routing for the packets
// It is responsible for routing packets to the appropriate handlers
// The implementation is based on the UDP protocol
// It Handles Packet Routing, Starts and Runs the Router
// Add callbacks to the Router to handle incoming packets
// It may contains some default callbacks required for handling the router state
package router

import (
	"errors"
	"sync"

	"github.com/aura-speak/networking/pkg/protocol"
)

// ClientPacketHandler is the function type for the ClientPacketHandler
// It takes a packet and returns an error
type ClientPacketHandler func(packet *protocol.Packet) error

// ClientPacketRouter is the main struct for the ClientPacketRouter
// It contains the handlers for the ClientPacketRouter
type ClientPacketRouter struct {
	handlers sync.Map // packetType -> PacketHandler
}

// NewClientPacketRouter creates a new PacketRouter
func NewClientPacketRouter() *ClientPacketRouter {
	return &ClientPacketRouter{
		handlers: sync.Map{},
	}
}

// OnPacket registers a new PacketHandler for a specific packet type
//
// Example:
//
//	router.OnPacket(protocol.PacketTypeDebugHello, func(packet *protocol.Packet) error {
//		fmt.Println("Received text packet:", string(packet))
//		return nil
//	})
func (r *ClientPacketRouter) OnPacket(packetType protocol.PacketType, handler ClientPacketHandler) {
	r.handlers.Store(packetType, handler)
}

// HandlePacket routes a packet to the appropriate handler
//
// Example:
//
// router.HandlePacket(protocol.PacketTypeDebugHello, []byte("Hello, Server!"))
//
// # It returns an error if no handler is found for the packet type
//
// Example:
//
// err := router.HandlePacket(protocol.PacketTypeDebugHello, []byte("Hello, Server!"))
//
//	if err != nil {
//		fmt.Println("Error routing packet:", err)
//	}
func (r *ClientPacketRouter) HandlePacket(packet *protocol.Packet) error {
	// Check if the packet type is valid
	if !protocol.IsValidPacketType(packet.PacketHeader.PacketType) {
		return errors.New("invalid packet type")
	}
	// Load the handler for the packet type
	handler, ok := r.handlers.Load(packet.PacketHeader.PacketType)
	if !ok {
		return errors.New("no handler found for packet type")
	}
	// Cast the handler to the PacketHandler type
	handlerFunc := handler.(ClientPacketHandler)
	// Call the handler function
	return handlerFunc(packet)
}
