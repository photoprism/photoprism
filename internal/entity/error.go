package entity

import (
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/txt/clip"
)

// ErrorMessageBytes is the storage limit of the error_message column in bytes,
// matching its VARBINARY(2048) definition. Messages may contain multi-byte text
// (e.g. file paths or recovered-panic stack traces), so writes are clipped by
// bytes on a rune boundary to avoid write errors and invalid UTF-8.
const ErrorMessageBytes = 2048

// logEvents is true when events are being recorded in the "errors" database table.
var logEvents = atomic.Bool{}

// LogWarningsAndErrors starts logging published error and warning
// events to the errors database table if a database instance is set.
func LogWarningsAndErrors() {
	if !HasDbProvider() {
		return
	}

	if logEvents.CompareAndSwap(false, true) {
		if err := mutex.ErrorWorker.Start(); err != nil {
			// This should not be possible, as the CompareAndSwap should prevent this scenario.
			return
		}
		go Error{}.LogEvents(logrus.WarnLevel)
	}
}

// Error represents an error message log.
type Error struct {
	ID           uint      `gorm:"primaryKey;" json:"ID" yaml:"ID"`
	ErrorTime    time.Time `gorm:"index" json:"Time" yaml:"Time"`
	ErrorLevel   string    `gorm:"type:bytes;size:32" json:"Level" yaml:"Level"`
	ErrorMessage string    `gorm:"type:bytes;size:2048" json:"Message" yaml:"Message"`
}

// Errors represents a list of error log messages.
type Errors []Error

// TableName returns the entity table name.
func (Error) TableName() string {
	return "errors"
}

// LogEvents writes published events with the specified minimum level to the "errors" database table.
func (Error) LogEvents(minLevel logrus.Level) {
	s := event.Subscribe("log.*")

	defer func() {
		mutex.ErrorWorker.Stop() // Make sure that the worker has been stopped, before swapping the state.
		logEvents.CompareAndSwap(true, false)
		event.Unsubscribe(s)
		mutex.ErrorWorker.Stop()
	}()

	// Wait for log events and write them to the  "errors" table,
	// as long as a database connection exists.
	for msg := range s.Receiver {
		if !mutex.ErrorWorker.Canceled() {
			var err error
			var level logrus.Level

			if val, ok := msg.Fields["level"]; !ok {
				continue
			} else if level, err = logrus.ParseLevel(val.(string)); err != nil || level > minLevel {
				continue
			}

			errLog := Error{ErrorLevel: level.String()}

			if val, ok := msg.Fields["message"]; ok {
				errLog.ErrorMessage = clip.Bytes(val.(string), ErrorMessageBytes)
			}

			if val, ok := msg.Fields["time"]; ok {
				errLog.ErrorTime = val.(time.Time)
			}

			if HasDbProvider() {
				Db().Create(&errLog)
			} else {
				return
			}
		} else {
			return
		}
	}
}
