package logger

import (
	"github.com/aura-speak/networking/internal/web"
	"github.com/sirupsen/logrus"
)

type WebSocketHook struct {
	hub       *web.WebSocketHub
	formatter logrus.Formatter
}

func (wh *WebSocketHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (wh *WebSocketHook) Fire(entry *logrus.Entry) error {
	e := entry.Dup()
	if wh.hub == nil {
		return nil
	}
	b, err := wh.formatter.Format(e)
	if err != nil {
		return err
	}
	wh.hub.Broadcast([]byte(b))
	return nil
}
