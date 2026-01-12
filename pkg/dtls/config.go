package dtls

import (
	"time"

	"github.com/aura-speak/networking/pkg/dtls/ciphersuite"
)

/*
Handshake-Logik: HandshakeTimeout, InitialRTO, MaxRTO, MaxRetransmits, CookieLifetime
Runtime/API: ReadTimeout, WriteTimeout, IdleTimeout, KeepaliveEvery
*/

const DEFAULT_PORT = "54321"

type Timeouts struct {
	// Overall Timeout for Handshake
	// no connection after that cancel whole Handshake
	HandshakeTimeout time.Duration // Defaults to 60s
	// The duration that is allowed from the first ClientHello to an answer
	InitialRTO time.Duration // Defaults to 500ms
	// Prevents to long retransmit wait times
	MaxRTO time.Duration // defaults to 60s
	// Most allowed retransmits before connection breaks
	MaxRetransmits int // defaults to 7

	// Cookie allowed lifetime
	CookieLifetime time.Duration // defaults to. 20s

	// How long is the conn.Read() allowed to block
	ReadTimeout time.Duration // defaults to 30s
	// How long is the conn.Write() allowed to block
	WriteTimeout time.Duration // defaults to 5s
	// Time that it only transmit non or non valid packages
	IdleTimeout time.Duration // defaults to 2m
	// Send Keep alive package interval
	KeepaliveEvery time.Duration // defaults to 25s
}

type Config struct {
	// Max Datagram Length
	MTU int // defaults to 1200

	Timeouts // all the timeouts
	// secret for the Cookie only server side keep client side empty
	CookieSecret     string           // The Cookie Secrets
	CipherSuits      []ciphersuite.ID // supported cipher suits
	ReplayWindowSize uint8            // defaults to 128
	EnableCID        bool             // Should CID be enabled?
}
