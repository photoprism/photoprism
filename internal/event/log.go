package event

import (
	"fmt"

	"github.com/leandro-lugaresi/hub"
	"github.com/sirupsen/logrus"
)

// TextFormatter for log messages.
var TextFormatter = &logrus.TextFormatter{
	DisableColors: false,
	FullTimestamp: true,
}

// Log is the global default logger.
var Log Logger
var LogBuffer Buffer

// Hook represents a log event hook.
type Hook struct {
	hub *hub.Hub
}

// NewHook creates a new log event hook.
func NewHook(hub *hub.Hub) *Hook {
	return &Hook{hub: hub}
}

// Fire publishes a new log event,
func (h *Hook) Fire(entry *logrus.Entry) error {
	if entry == nil {
		return fmt.Errorf("log entry is empty")
	} else if entry.Message == "" {
		return fmt.Errorf("log message is empty")
	} else if LogBuffer.Get() == entry.Message {
		return nil
	} else if err := LogBuffer.Set(entry.Message); err != nil {
		return err
	}

	h.hub.Publish(Message{
		Name: "log." + entry.Level.String(),
		Fields: Data{
			"time":    entry.Time,
			"level":   entry.Level.String(),
			"message": entry.Message,
		},
	})

	return nil
}

// Levels returns a slice containing all supported log levels.
func (h *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}
