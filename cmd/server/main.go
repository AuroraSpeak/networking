package main

import (
	"context"

	"github.com/aura-speak/networking/internal/config"
	"github.com/aura-speak/networking/pkg/protocol"
	"github.com/aura-speak/networking/pkg/server"
)

func main() {
	ctx := context.Background()
	cfg := config.ServerConfigLoader()
	server := server.NewServer(8080, ctx, cfg)
	server.OnPacket(protocol.PacketTypeDebugAny, func(packet *protocol.Packet, clientAddr string) error {
		server.Broadcast(packet)
		return nil
	})
	server.Run()
}
