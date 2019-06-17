package models

import (
	"time"

	"github.com/jinzhu/gorm"

	uuid "github.com/satori/go.uuid"
)

// Shared photos and/or albums
type Share struct {
	ShareUUID     string `gorm:"primary_key;auto_increment:false"`
	PhotoID       uint
	AlbumID       uint
	ShareViews    uint
	ShareSecret   string
	SharePassword string
	ShareExpires  time.Time
	Photo         *Photo
	Album         *Album
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time `sql:"index"`
}

func (Share) TableName() string {
	return "shares"
}

func (s *Share) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("ShareUUID", uuid.NewV4().String())
}
