package dtls

import (
	"net"
	"time"
)

type inboundDatagram struct {
	data []byte
	addr *net.UDPAddr
	ts   time.Time
}
