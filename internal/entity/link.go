package entity

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Link represents a sharing link.
type Link struct {
	LinkToken    string     `gorm:"type:varbinary(255);primary_key;"`
	LinkPassword string     `gorm:"type:varbinary(255);"`
	LinkExpires  *time.Time `gorm:"type:datetime;"`
	ShareUUID    string     `gorm:"type:varbinary(36);index;"`
	CanComment   bool
	CanEdit      bool
	CreatedAt    time.Time  `deepcopier:"skip"`
	UpdatedAt    time.Time  `deepcopier:"skip"`
	DeletedAt    *time.Time `deepcopier:"skip" sql:"index"`
}

// BeforeCreate creates a new URL token when a new link is created.
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
