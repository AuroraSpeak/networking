package main

import (
	"context"

	"github.com/aura-speak/networking/pkg/protocol"
	"github.com/aura-speak/networking/pkg/server"
)

func main() {
	ctx := context.Background()
	server := server.NewServer(8080, ctx)
	server.OnPacket(protocol.PacketTypeDebugAny, func(packet *protocol.Packet, clientAddr string) error {
		server.Broadcast(packet)
		return nil
	})
	server.Run()
}
