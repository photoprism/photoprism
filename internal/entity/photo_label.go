package entity

import (
	"errors"

	"github.com/photoprism/photoprism/internal/ai/classify"
)

type PhotoLabels []PhotoLabel

// PhotoLabel represents the many-to-many relation between Photo and label.
// Labels are weighted by uncertainty (100 - confidence)
type PhotoLabel struct {
	PhotoID     uint   `gorm:"primaryKey;autoIncrement:false"`
	LabelID     uint   `gorm:"primaryKey;autoIncrement:false"`
	LabelSrc    string `gorm:"type:bytes;size:8;"`
	Uncertainty int    `gorm:"type:SMALLINT"`
	Photo       *Photo `gorm:"foreignKey:PhotoID;references:ID;" yaml:"-"`
	Label       *Label `gorm:"foreignKey:LabelID;references:ID;"`
}

// TableName returns the entity table name.
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

// Updates multiple columns in the database.
func (m *PhotoLabel) Updates(values interface{}) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// Update a column in the database.
func (m *PhotoLabel) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *PhotoLabel) Save() error {
	if m.Photo != nil {
		if m.PhotoID == 0 {
			m.PhotoID = m.Photo.ID
		}
		m.Photo = nil
	}
	if m.PhotoID == 0 {
		return errors.New("PK value not provided")
	}

	if m.Label != nil {
		m.Label.SetName(m.Label.LabelName)
		if err := m.Label.Save(); err != nil {
			return err
		}
	}

	return Db().Save(m).Error
}

// Create inserts a new row to the database.
func (m *PhotoLabel) Create() error {
	return Db().Create(m).Error
}

// Delete deletes the label reference.
func (m *PhotoLabel) Delete() error {
	return Db().Delete(m).Error
}

// FirstOrCreatePhotoLabel returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreatePhotoLabel(m *PhotoLabel) *PhotoLabel {
	result := PhotoLabel{}

	if err := Db().Where("photo_id = ? AND label_id = ?", m.PhotoID, m.LabelID).First(&result).Error; err == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if err := Db().Where("photo_id = ? AND label_id = ?", m.PhotoID, m.LabelID).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("photo-label: %s (find or create)", createErr)
	}

	return nil
}

// ClassifyLabel returns the label as classify.Label
func (m *PhotoLabel) ClassifyLabel() classify.Label {
	if m.Label == nil {
		log.Errorf("photo-label: classify label is nil (photo id %d, label id %d) - you may have found a bug", m.PhotoID, m.LabelID)
		return classify.Label{}
	}

	result := classify.Label{
		Name:        m.Label.LabelName,
		Source:      m.LabelSrc,
		Uncertainty: m.Uncertainty,
		Priority:    m.Label.LabelPriority,
	}

	return result
}
