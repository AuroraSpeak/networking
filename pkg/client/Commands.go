package client

type InternalCommand int

// Internal commands for client operations
const (
	// update udp client state
	CmdUpdateClientState InternalCommand = iota
)
