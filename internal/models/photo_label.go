package models

import (
	"github.com/jinzhu/gorm"
)

// Photo labels are weighted by uncertainty (100 - confidence)
type PhotoLabel struct {
	PhotoID          uint `gorm:"primary_key;auto_increment:false"`
	LabelID          uint `gorm:"primary_key;auto_increment:false;index"`
	LabelUncertainty int
	LabelSource      string
	Photo            *Photo
	Label            *Label
}

func (PhotoLabel) TableName() string {
	return "photos_labels"
}

func NewPhotoLabel(photoId, labelId uint, uncertainty int, source string) *PhotoLabel {
	result := &PhotoLabel{
		PhotoID:          photoId,
		LabelID:          labelId,
		LabelUncertainty: uncertainty,
		LabelSource:      source,
	}

	return result
}

func (m *PhotoLabel) FirstOrCreate(db *gorm.DB) *PhotoLabel {
	db.FirstOrCreate(m, "photo_id = ? AND label_id = ?", m.PhotoID, m.LabelID)

	return m
}
