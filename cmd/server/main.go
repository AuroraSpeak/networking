package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/aura-speak/networking/pkg/protocol"
	"github.com/aura-speak/networking/pkg/server"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := server.NewServer(8080, ctx)
	srv.OnPacket(protocol.PacketTypeDebugAny, func(packet *protocol.Packet, clientAddr string) error {
		srv.Broadcast(packet)
		return nil
	})

	// Signal Handler für graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Info("Shutting down server...")
		cancel()
		srv.Stop()
	}()

	if err := srv.Run(); err != nil {
		log.WithError(err).Error("Server error")
	}
}
