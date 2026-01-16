//go:build !debug
// +build !debug

package adapter

import (
	"github.com/aura-speak/networking/internal/web/orchestrator"
	"github.com/aura-speak/networking/pkg/client"
)

// DefaultClientFactory returns the release client factory
func DefaultClientFactory() orchestrator.ClientFactory {
	return func(host string, port int, id int) *client.Client {
		return client.NewClient(host, port)
	}
}
