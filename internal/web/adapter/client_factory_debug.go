//go:build debug
// +build debug

package adapter

import (
	"github.com/aura-speak/networking/internal/web/orchestrator"
	"github.com/aura-speak/networking/pkg/client"
)

// DefaultClientFactory returns the debug client factory
func DefaultClientFactory() orchestrator.ClientFactory {
	return func(host string, port int, id int) *client.Client {
		return client.NewDebugClient(host, port, id)
	}
}
