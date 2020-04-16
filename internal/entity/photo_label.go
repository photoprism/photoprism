package entity

import (
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/mutex"
)

// PhotoLabel represents the many-to-many relation between Photo and label.
// Labels are weighted by uncertainty (100 - confidence)
type PhotoLabel struct {
	PhotoID          uint `gorm:"primary_key;auto_increment:false"`
	LabelID          uint `gorm:"primary_key;auto_increment:false;index"`
	LabelUncertainty int
	LabelSource      string
	Photo            *Photo
	Label            *Label
}

// TableName returns PhotoLabel table identifier "photos_labels"
func (PhotoLabel) TableName() string {
	return "photos_labels"
}

// NewPhotoLabel registers a new PhotoLabel relation with an uncertainty and a source of label
func NewPhotoLabel(photoID, labelID uint, uncertainty int, source string) *PhotoLabel {
	result := &PhotoLabel{
		PhotoID:          photoID,
		LabelID:          labelID,
		LabelUncertainty: uncertainty,
		LabelSource:      source,
	}

	return result
}

// FirstOrCreate checks wether the PhotoLabel relation already exist in the database before the creation
func (m *PhotoLabel) FirstOrCreate(db *gorm.DB) *PhotoLabel {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "photo_id = ? AND label_id = ?", m.PhotoID, m.LabelID).Error; err != nil {
		log.Errorf("photo label: %s", err)
	}

	return m
}

// ClassifyLabel returns the label as classify.Label
func (m *PhotoLabel) ClassifyLabel() classify.Label {
	if m.Label == nil {
		panic("photo label: label is nil")
	}

	result := classify.Label{
		Name: m.Label.LabelName,
		Source: m.LabelSource,
		Uncertainty: m.LabelUncertainty,
		Priority: m.Label.LabelPriority,
	}

	return result
}
