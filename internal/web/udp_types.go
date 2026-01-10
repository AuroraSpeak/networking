package web

import (
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type datagramDirection int

const (
	ClientToServer = 1
	ServerToClient = 2
)

type datagram struct {
	Direction datagramDirection `json:"direction"`
	Message   []byte            `json:"message"`
}

func (d *datagram) Send(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(d)
	if err != nil {
		log.WithField("caller", "web").WithError(err).Error("Can't marshal Datagram to json")
	}
	w.Write(b)
	w.Write([]byte("\n"))
}

type trace struct {
	TS time.Time `json:"ts"`
}
