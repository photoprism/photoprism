package entity

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/util"

	uuid "github.com/satori/go.uuid"
)

// Shared photos and/or albums
type Share struct {
	ShareUUID     string `gorm:"primary_key;auto_increment:false"`
	PhotoUUID     string
	AlbumUUID     string
	ShareViews    uint
	ShareUrl      string `gorm:"type:varchar(64);"`
	ShareToken    string `gorm:"type:varchar(64);"`
	SharePassword string `gorm:"type:varchar(128);"`
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
	if err := scope.SetColumn("ShareUUID", uuid.NewV4().String()); err != nil {
		return err
	}

	if err := scope.SetColumn("ShareToken", util.RandomToken(4)); err != nil {
		return err
	}

	return nil
}
