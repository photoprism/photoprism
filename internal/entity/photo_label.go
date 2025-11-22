package entity

import (
	"errors"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/ai/classify"
)

// PhotoLabels is a convenience alias for lists of PhotoLabel relations.
type PhotoLabels []PhotoLabel

// PhotoLabel represents the many-to-many relation between Photo and Label.
// Labels are weighted by uncertainty (100 - confidence).
type PhotoLabel struct {
	PhotoID     uint   `gorm:"primary_key;auto_increment:false" json:"PhotoID,omitempty" yaml:"PhotoID"`
	LabelID     uint   `gorm:"primary_key;auto_increment:false;index" json:"LabelID,omitempty" yaml:"LabelID"`
	LabelSrc    string `gorm:"type:VARBINARY(8);" json:"LabelSrc,omitempty" yaml:"LabelSrc,omitempty"`
	Uncertainty int    `gorm:"type:SMALLINT" json:"Uncertainty" yaml:"Uncertainty"`
	Topicality  int    `gorm:"type:SMALLINT;default:0;" json:"Topicality" yaml:"Topicality,omitempty"`
	NSFW        int    `gorm:"type:SMALLINT;column:nsfw;default:0;" json:"NSFW,omitempty" yaml:"NSFW,omitempty"`
	Photo       *Photo `gorm:"PRELOAD:false" json:"-" yaml:"-"`
	Label       *Label `gorm:"PRELOAD:true" json:"Label,omitempty" yaml:"-"`
}

// TableName returns the database table name for PhotoLabel.
func (PhotoLabel) TableName() string {
	return "photos_labels"
}

// NewPhotoLabel registers a new PhotoLabel relation with an uncertainty and source.
func NewPhotoLabel(photoID, labelID uint, uncertainty int, source string) *PhotoLabel {
	result := &PhotoLabel{
		PhotoID:     photoID,
		LabelID:     labelID,
		Uncertainty: uncertainty,
		LabelSrc:    source,
	}

	return result
}

// Updates mutates multiple columns in the database and clears cached copies.
func (m *PhotoLabel) Updates(values interface{}) error {
	if m == nil {
		return errors.New("photo label must not be nil - you may have found a bug")
	} else if !m.HasID() {
		return errors.New("photo label ID must not be empty - you may have found a bug")
	}

	if err := UnscopedDb().Model(m).UpdateColumns(values).Error; err != nil {
		return err
	}

	FlushCachedPhotoLabel(m)
	return nil
}

// Update mutates a single column in the database and clears cached copies.
func (m *PhotoLabel) Update(attr string, value interface{}) error {
	if m == nil {
		return errors.New("photo label must not be nil - you may have found a bug")
	} else if !m.HasID() {
		return errors.New("photo label ID must not be empty - you may have found a bug")
	}

	if err := UnscopedDb().Model(m).UpdateColumn(attr, value).Error; err != nil {
		return err
	}

	FlushCachedPhotoLabel(m)
	return nil
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *PhotoLabel) Save() error {
	if m.Photo != nil {
		// Clear the eager-loaded Photo pointer so GORM does not attempt to persist it again.
		m.Photo = nil
	}

	if m.Label == nil {
		// Do nothing.
	} else if !m.Label.SetName(m.Label.LabelName) {
		return ErrInvalidName
	}

	return Db().Save(m).Error
}

// Create inserts a new row into the database without touching cache state.
func (m *PhotoLabel) Create() error {
	FlushCachedPhotoLabel(m)
	return Db().Create(m).Error
}

// AfterCreate flushes the label cache once a relation has been persisted.
func (m *PhotoLabel) AfterCreate(scope *gorm.Scope) error {
	FlushCachedPhotoLabel(m)
	return nil
}

// AfterUpdate flushes the label cache after a relation change.
func (m *PhotoLabel) AfterUpdate(tx *gorm.DB) (err error) {
	FlushCachedPhotoLabel(m)
	return
}

// Delete removes the label reference and clears the cache.
func (m *PhotoLabel) Delete() error {
	FlushCachedPhotoLabel(m)
	return Db().Delete(m).Error
}

// AfterDelete flushes the label cache when a label is deleted.
func (m *PhotoLabel) AfterDelete(tx *gorm.DB) (err error) {
	FlushCachedPhotoLabel(m)
	return
}

// HasID tests if both a photo and label ID are set.
func (m *PhotoLabel) HasID() bool {
	if m == nil {
		return false
	}

	return m.PhotoID > 0 && m.LabelID > 0
}

// CacheKey returns a string key for caching the entity.
func (m *PhotoLabel) CacheKey() string {
	return photoLabelCacheKey(m.PhotoID, m.LabelID)
}

// FirstOrCreatePhotoLabel returns the existing row, inserts a new row, or nil in case of errors.
func FirstOrCreatePhotoLabel(m *PhotoLabel) *PhotoLabel {
	if m == nil {
		return nil
	} else if !m.HasID() {
		return nil
	}

	// Try to find and return an existing label. Otherwise, create a new one and return it.
	if result, err := FindPhotoLabel(m.PhotoID, m.LabelID, true); err == nil {
		return result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if result, err = FindPhotoLabel(m.PhotoID, m.LabelID, false); err == nil {
		return result
	} else {
		log.Errorf("photo-label: %s (find or create)", createErr)
	}

	return nil
}

// ClassifyLabel returns the label as a classify.Label.
func (m *PhotoLabel) ClassifyLabel() classify.Label {
	if m.Label == nil {
		log.Errorf("photo-label: classify label is nil (photo id %d, label id %d) - you may have found a bug", m.PhotoID, m.LabelID)
		return classify.Label{}
	}

	result := classify.Label{
		Name:           m.Label.LabelName,
		Source:         m.LabelSrc,
		Uncertainty:    m.Uncertainty,
		Topicality:     m.Topicality,
		Priority:       m.Label.LabelPriority,
		NSFW:           m.Label.LabelNSFW,
		NSFWConfidence: m.NSFW,
	}

	return result
}
