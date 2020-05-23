package entity

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Link represents a sharing link.
type Link struct {
	LinkToken    string     `gorm:"type:varbinary(255);primary_key;" json:"Token"`
	LinkPassword string     `gorm:"type:varbinary(255);" json:"Password"`
	LinkExpires  *time.Time `gorm:"type:datetime;" json:"Expires"`
	ShareUID     string     `gorm:"type:varbinary(36);index;" json:"ShareUID"`
	CanComment   bool       `json:"CanComment"`
	CanEdit      bool       `json:"CanEdit"`
	CreatedAt    time.Time  `deepcopier:"skip" json:"CreatedAt"`
	UpdatedAt    time.Time  `deepcopier:"skip" json:"UpdatedAt"`
	DeletedAt    *time.Time `deepcopier:"skip" sql:"index" json:"DeletedAt,omitempty"`
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Link) BeforeCreate(scope *gorm.Scope) error {
	if err := scope.SetColumn("LinkToken", rnd.Token(10)); err != nil {
		return err
	}

	return nil
}

// NewLink creates a sharing link.
func NewLink(password string, canComment, canEdit bool) Link {
	result := Link{
		LinkToken:    rnd.Token(10),
		LinkPassword: password,
		CanComment:   canComment,
		CanEdit:      canEdit,
	}

	return result
}
