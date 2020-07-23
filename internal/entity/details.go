package entity

import (
	"fmt"
	"time"
)

// Details stores additional metadata fields for each photo to improve search performance.
type Details struct {
	PhotoID   uint      `gorm:"primary_key;auto_increment:false" yaml:"-"`
	Keywords  string    `gorm:"type:text;" json:"Keywords" yaml:"Keywords"`
	Notes     string    `gorm:"type:text;" json:"Notes" yaml:"Notes,omitempty"`
	Subject   string    `gorm:"type:varchar(255);" json:"Subject" yaml:"Subject,omitempty"`
	Artist    string    `gorm:"type:varchar(255);" json:"Artist" yaml:"Artist,omitempty"`
	Copyright string    `gorm:"type:varchar(255);" json:"Copyright" yaml:"Copyright,omitempty"`
	License   string    `gorm:"type:varchar(255);" json:"License" yaml:"License,omitempty"`
	CreatedAt time.Time `yaml:"-"`
	UpdatedAt time.Time `yaml:"-"`
}

// NewDetails creates new photo details.
func NewDetails(photo Photo) Details {
	return Details{PhotoID: photo.ID}
}

// Create inserts a new row to the database.
func (m *Details) Create() error {
	if m.PhotoID == 0 {
		return fmt.Errorf("details: photo id must not be empty (create)")
	}

	return UnscopedDb().Create(m).Error
}

// Save updates existing photo details or inserts a new row.
func (m *Details) Save() error {
	if m.PhotoID == 0 {
		return fmt.Errorf("details: photo id must not be empty (save)")
	}

	return UnscopedDb().Save(m).Error
}

// FirstOrCreateDetails returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateDetails(m *Details) *Details {
	result := Details{}

	if err := m.Create(); err == nil {
		return m
	} else if err := Db().Where("photo_id = ?", m.PhotoID).First(&result).Error; err == nil {
		if m.CreatedAt.IsZero() {
			m.CreatedAt = Timestamp()
		}

		return &result
	} else {
		log.Errorf("details: %s (first or create %d)", err, m.PhotoID)
	}

	return nil
}

// NoKeywords checks if the photo has no Keywords
func (m *Details) NoKeywords() bool {
	return m.Keywords == ""
}

// NoSubject checks if the photo has no Subject
func (m *Details) NoSubject() bool {
	return m.Subject == ""
}

// NoNotes checks if the photo has no Notes
func (m *Details) NoNotes() bool {
	return m.Notes == ""
}

// NoArtist checks if the photo has no Artist
func (m *Details) NoArtist() bool {
	return m.Artist == ""
}

// NoCopyright checks if the photo has no Copyright
func (m *Details) NoCopyright() bool {
	return m.Copyright == ""
}
