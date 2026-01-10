package router

import (
	"errors"
	"fmt"
	"sync"

	"github.com/aura-speak/networking/pkg/protocol"
)

type ServerPacketHandler func(packet *protocol.Packet, clientAddr string) error

type ServerPacketRouter struct {
	handlers sync.Map // packetType -> PacketHandler
}

// NewServerPacketRouter creates a new ServerPacketRouter
func NewServerPacketRouter() *ServerPacketRouter {
	return &ServerPacketRouter{
		handlers: sync.Map{},
	}
}

// OnPacket registers a new PacketHandler for a specific packet type
// Example:
//
//	router.OnPacket(protocol.PacketTypeDebugHello, func(packet *protocol.Packet, clientAddr string) error {
//		fmt.Println("Received debug hello packet from client:", clientAddr)
//		return nil
//	})
func (r *ServerPacketRouter) OnPacket(packetType protocol.PacketType, handler ServerPacketHandler) {
	r.handlers.Store(packetType, handler)
}

// HandlePacket handles a packet from a client
// Example:
//
//	router.HandlePacket(packet, clientAddr)
//	if err != nil {
//		fmt.Println("Error handling packet:", err)
//	}
func (r *ServerPacketRouter) HandlePacket(packet *protocol.Packet, clientAddr string) error {
	if !protocol.IsValidPacketType(packet.PacketHeader.PacketType) {
		return errors.New("invalid packet type")
	}
	handler, ok := r.handlers.Load(packet.PacketHeader.PacketType)
	if !ok {
		strPacketType, exists := protocol.PacketTypeMapType[packet.PacketHeader.PacketType]
		if !exists {
			strPacketType = fmt.Sprintf("Unknown(0x%02X)", packet.PacketHeader.PacketType)
		}
		return fmt.Errorf("no handler found for packet type: %s", strPacketType)
	}
	handlerFunc := handler.(ServerPacketHandler)
	return handlerFunc(packet, clientAddr)
}

func (r *ServerPacketRouter) ListRoutes() {
	r.handlers.Range(func(key, value interface{}) bool {
		strPacketType, exists := protocol.PacketTypeMapType[key.(protocol.PacketType)]
		if !exists {
			strPacketType = fmt.Sprintf("Unknown(0x%02X)", key.(protocol.PacketType))
		}
		fmt.Printf("Packet type: %s\n", strPacketType)
		return true
	})
}
