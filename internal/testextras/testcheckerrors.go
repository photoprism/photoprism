package testextras

import (
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"gorm.io/gorm"
)

// Error represents an error message log.
type Error struct {
	ID           uint      `gorm:"primaryKey;" json:"ID" yaml:"ID"`
	ErrorTime    time.Time `sql:"index" json:"Time" yaml:"Time"`
	ErrorLevel   string    `gorm:"type:bytes;size:32" json:"Level" yaml:"Level"`
	ErrorMessage string    `gorm:"type:bytes;size:2048" json:"Message" yaml:"Message"`
}

// Errors represents a list of error log messages.
type Errors []Error

// TableName returns the entity table name.
func (Error) TableName() string {
	return "errors"
}

func ValidateDBErrors(db *gorm.DB, log event.Logger, beforeTimestamp time.Time, code int) int {
	errorMessage := "%threw photo.Save has inconsistent%"
	var afterErrors Errors
	db.Where("error_time > ? AND error_message LIKE ?", beforeTimestamp, errorMessage).Find(&afterErrors)

	if len(afterErrors) > 0 {
		code = 1 // Force the test suite as failed as unexpected Save errors have been stored in the database.
		for _, element := range afterErrors {
			log.Errorf("Save Error found %v", element)
		}
	}
	return code
}
