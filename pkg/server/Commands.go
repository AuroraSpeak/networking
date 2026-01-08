package server

type InternalCommand int

// Internal commands for server operations
const (
	// update udp server state
	CmdUpdateServerState InternalCommand = iota
)
