package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aura-speak/networking/internal/logger"
	"github.com/aura-speak/networking/internal/web/adapter"
	log "github.com/sirupsen/logrus"
)

func main() {
	logger.Setup()
	server := adapter.NewServer(8080, 9090, adapter.DefaultClientFactory())

	// Starte Server in Goroutine
	go func() {
		if err := server.Run(); err != nil {
			log.WithError(err).Error("error starting server")
			// Ignoriere ErrServerClosed, das ist normal beim Shutdown
			if err.Error() != "http: Server closed" {
				panic(err)
			}
		}
	}()

	// Warte auf Shutdown-Signale
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	println("\nReceived shutdown signal")

	// Graceful shutdown mit 10 Sekunden Timeout
	if err := server.Shutdown(10 * time.Second); err != nil {
		panic(err)
	}
}
