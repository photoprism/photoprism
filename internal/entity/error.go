package entity

import (
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/sirupsen/logrus"
)

// Error represents an error message log.
type Error struct {
	ID           uint      `gorm:"primary_key"`
	ErrorTime    time.Time `sql:"index"`
	ErrorLevel   string    `gorm:"type:varbinary(32)"`
	ErrorMessage string    `gorm:"type:varbinary(2048)"`
}

// SaveErrorMessages subscribes to error logs and stored them in the errors table.
func SaveErrorMessages() {
	s := event.Subscribe("log.*")

	defer func() {
		event.Unsubscribe(s)
	}()

	for msg := range s.Receiver {
		level, ok := msg.Fields["level"]

		if !ok {
			continue
		}

		logLevel, err := logrus.ParseLevel(level.(string))

		if err != nil || logLevel >= logrus.InfoLevel {
			continue
		}

		newError := Error{ErrorLevel: logLevel.String()}

		if val, ok := msg.Fields["msg"]; ok {
			newError.ErrorMessage = val.(string)
		}

		if val, ok := msg.Fields["time"]; ok {
			newError.ErrorTime = val.(time.Time)
		}

		Db().Create(&newError)
	}
}
