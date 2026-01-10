//go:build !debug
// +build !debug

package server

import "github.com/aura-speak/networking/pkg/protocol"

func tryRegisterClient(remote string, id int) (registered bool) { return false }

func lookupClientID(remote string) (int, bool) { return 0, false }

func (s *Server) handleDebugHello(packet *protocol.Packet, clientAddr string) error { return nil }
