package entity

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Shared photos and/or albums
type Share struct {
	UUID          string `gorm:"type:varbinary(36);primary_key;auto_increment:false"`
	ShareUUID     string `gorm:"type:varbinary(36);index;"`
	ShareViews    uint
	ShareUrl      string `gorm:"type:varchar(64);"`
	SharePassword string `gorm:"type:varbinary(200);"`
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
	if err := scope.SetColumn("ShareUUID", ID('s')); err != nil {
		return err
	}

	return nil
}
