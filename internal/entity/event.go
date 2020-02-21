package entity

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Event defines temporal event that can be used to link photos together
type Event struct {
	EventUUID        string `gorm:"type:varbinary(36);unique_index;"`
	EventSlug        string `gorm:"type:varbinary(128);unique_index;"`
	EventName        string
	EventType        string
	EventDescription string    `gorm:"type:text;"`
	EventNotes       string    `gorm:"type:text;"`
	EventBegin       time.Time `gorm:"type:datetime;"`
	EventEnd         time.Time `gorm:"type:datetime;"`
	EventLat         float64
	EventLng         float64
	EventDist        float64
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time `sql:"index"`
}

// TableName returns Event table identifier "events"
func (Event) TableName() string {
	return "events"
}

// BeforeCreate computes a random UUID when a new event is created in database
func (e *Event) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("EventUUID", rnd.PPID('e'))
}
