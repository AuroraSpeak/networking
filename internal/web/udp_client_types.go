package web

import (
	"github.com/aura-speak/networking/pkg/client"
)

type udpClientAction int

const (
	UDPClientActionAddDatagram udpClientAction = iota
)

type udpClient struct {
	// ID of the client
	id int
	// The UDP client
	client *client.Client
	// name random chosen
	name string
	// Datagram by the user
	datagrams []datagram
	// is it running
	running bool
}
type UDPClientActionData struct {
	id     int
	Action udpClientAction
}
