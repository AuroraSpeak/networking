package main

import (
	"context"

	"github.com/aura-speak/networking/pkg/server"
)

func main() {
	ctx := context.Background()
	server := server.NewServer(8080, ctx)
	server.OnPacket("", func(packet []byte) error {
		server.Broadcast(packet)
		return nil
	})
	server.Run()
}
