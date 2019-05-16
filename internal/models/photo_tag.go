package models

import (
	"github.com/jinzhu/gorm"
)

// Photo tags are weighted by confidence (probability * 100)
type PhotoTag struct {
	PhotoID       uint `gorm:"primary_key"`
	TagID         uint `gorm:"primary_key"`
	TagConfidence uint
	TagSource     string
}

func (PhotoTag) TableName() string {
	return "photo_tags"
}

func (t *PhotoTag) FirstOrCreate(db *gorm.DB) *PhotoTag {
	db.FirstOrCreate(t, "photo_id = ? AND tag_id = ?", t.PhotoID, t.TagID)

	return t
}
