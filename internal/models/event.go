package models

import (
	"time"

	"github.com/jinzhu/gorm"

	uuid "github.com/satori/go.uuid"
)

// Events
type Event struct {
	EventUUID        string `gorm:"primary_key;auto_increment:false"`
	EventSlug        string
	EventName        string
	EventType        string
	EventDescription string `gorm:"type:text;"`
	EventNotes       string `gorm:"type:text;"`
	EventBegin       time.Time
	EventEnd         time.Time
	EventLat         float64
	EventLong        float64
	EventDist        float64
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time `sql:"index"`
}

func (Event) TableName() string {
	return "events"
}

func (e *Event) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("EventUUID", uuid.NewV4().String())
}
