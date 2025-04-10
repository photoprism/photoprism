package entity

import (
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/event"
)

// logEvents is true when events are being recorded in the "errors" database table.
var logEvents = atomic.Bool{}

// LogWarningsAndErrors starts logging published error and warning
// events to the errors database table if a database instance is set.
func LogWarningsAndErrors() {
	if !HasDbProvider() {
		return
	}

	if logEvents.CompareAndSwap(false, true) {
		go Error{}.LogEvents(logrus.WarnLevel)
	}
}

// Error represents an error message log.
type Error struct {
	ID           uint      `gorm:"primary_key" json:"ID" yaml:"ID"`
	ErrorTime    time.Time `sql:"index" json:"Time" yaml:"Time"`
	ErrorLevel   string    `gorm:"type:VARBINARY(32)" json:"Level" yaml:"Level"`
	ErrorMessage string    `gorm:"type:VARBINARY(2048)" json:"Message" yaml:"Message"`
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
		logEvents.CompareAndSwap(true, false)
		event.Unsubscribe(s)
	}()

	// Wait for log events and write them to the  "errors" table,
	// as long as a database connection exists.
	for msg := range s.Receiver {
		var err error
		var level logrus.Level

		if val, ok := msg.Fields["level"]; !ok {
			continue
		} else if level, err = logrus.ParseLevel(val.(string)); err != nil || level > minLevel {
			continue
		}

		errLog := Error{ErrorLevel: level.String()}

		if val, ok := msg.Fields["message"]; ok {
			errLog.ErrorMessage = val.(string)
		}

		if val, ok := msg.Fields["time"]; ok {
			errLog.ErrorTime = val.(time.Time)
		}

		if HasDbProvider() {
			Db().Create(&errLog)
		} else {
			break
		}
	}
}
