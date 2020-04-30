package entity

import (
	"github.com/photoprism/photoprism/internal/classify"
)

// PhotoLabel represents the many-to-many relation between Photo and label.
// Labels are weighted by uncertainty (100 - confidence)
type PhotoLabel struct {
	PhotoID     uint   `gorm:"primary_key;auto_increment:false"`
	LabelID     uint   `gorm:"primary_key;auto_increment:false;index"`
	LabelSrc    string `gorm:"type:varbinary(8);"`
	Uncertainty int    `gorm:"type:SMALLINT"`
	Photo       *Photo `gorm:"PRELOAD:false"`
	Label       *Label `gorm:"PRELOAD:true"`
}

// TableName returns PhotoLabel table identifier "photos_labels"
func (PhotoLabel) TableName() string {
	return "photos_labels"
}

// NewPhotoLabel registers a new PhotoLabel relation with an uncertainty and a source of label
func NewPhotoLabel(photoID, labelID uint, uncertainty int, source string) *PhotoLabel {
	result := &PhotoLabel{
		PhotoID:     photoID,
		LabelID:     labelID,
		Uncertainty: uncertainty,
		LabelSrc:    source,
	}

	return result
}

// FirstOrCreate checks if the PhotoLabel relation already exist in the database before the creation
func (m *PhotoLabel) FirstOrCreate() *PhotoLabel {
	if err := Db().FirstOrCreate(m, "photo_id = ? AND label_id = ?", m.PhotoID, m.LabelID).Error; err != nil {
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
		Name:        m.Label.LabelName,
		Source:      m.LabelSrc,
		Uncertainty: m.Uncertainty,
		Priority:    m.Label.LabelPriority,
	}

	return result
}

// Save saves the entity in the database and returns an error.
func (m *PhotoLabel) Save() error {
	if m.Photo != nil {
		m.Photo = nil
	}

	if m.Label != nil {
		m.Label.SetName(m.Label.LabelName)
	}

	return Db().Save(m).Error
}
