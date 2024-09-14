package entity

import (
	"fmt"
	"sync"
	"time"

	"github.com/photoprism/photoprism/pkg/txt"
)

var photoDetailsMutex = sync.Mutex{}

// Details stores additional metadata fields for each photo to improve search performance.
type Details struct {
	PhotoID      uint      `gorm:"primary_key;auto_increment:false" yaml:"-"`
	Keywords     string    `gorm:"type:VARCHAR(2048);" json:"Keywords" yaml:"Keywords"`
	KeywordsSrc  string    `gorm:"type:VARBINARY(8);" json:"KeywordsSrc" yaml:"KeywordsSrc,omitempty"`
	Notes        string    `gorm:"type:VARCHAR(2048);" json:"Notes" yaml:"Notes,omitempty"`
	NotesSrc     string    `gorm:"type:VARBINARY(8);" json:"NotesSrc" yaml:"NotesSrc,omitempty"`
	Subject      string    `gorm:"type:VARCHAR(1024);" json:"Subject" yaml:"Subject,omitempty"`
	SubjectSrc   string    `gorm:"type:VARBINARY(8);" json:"SubjectSrc" yaml:"SubjectSrc,omitempty"`
	Artist       string    `gorm:"type:VARCHAR(1024);" json:"Artist" yaml:"Artist,omitempty"`
	ArtistSrc    string    `gorm:"type:VARBINARY(8);" json:"ArtistSrc" yaml:"ArtistSrc,omitempty"`
	Copyright    string    `gorm:"type:VARCHAR(1024);" json:"Copyright" yaml:"Copyright,omitempty"`
	CopyrightSrc string    `gorm:"type:VARBINARY(8);" json:"CopyrightSrc" yaml:"CopyrightSrc,omitempty"`
	License      string    `gorm:"type:VARCHAR(1024);" json:"License" yaml:"License,omitempty"`
	LicenseSrc   string    `gorm:"type:VARBINARY(8);" json:"LicenseSrc" yaml:"LicenseSrc,omitempty"`
	Software     string    `gorm:"type:VARCHAR(1024);" json:"Software" yaml:"Software,omitempty"`
	SoftwareSrc  string    `gorm:"type:VARBINARY(8);" json:"SoftwareSrc" yaml:"SoftwareSrc,omitempty"`
	CreatedAt    time.Time `yaml:"-"`
	UpdatedAt    time.Time `yaml:"-"`
}

// TableName returns the entity table name.
func (Details) TableName() string {
	return "details"
}

// NewDetails creates new photo details.
func NewDetails(photo Photo) Details {
	return Details{PhotoID: photo.ID}
}

// Create inserts a new row to the database.
func (m *Details) Create() error {
	photoDetailsMutex.Lock()
	defer photoDetailsMutex.Unlock()

	if m.PhotoID == 0 {
		return fmt.Errorf("details: photo id must not be empty (create)")
	}

	return UnscopedDb().Create(m).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
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
			m.CreatedAt = Now()
		}

		return &result
	} else {
		log.Errorf("details: %s (find or create %d)", err, m.PhotoID)
	}

	return nil
}

// NoKeywords tests if the photo has no Keywords.
func (m *Details) NoKeywords() bool {
	return m.Keywords == ""
}

// NoSubject tests if the photo has no Subject.
func (m *Details) NoSubject() bool {
	return m.Subject == ""
}

// NoNotes tests if the photo has no Notes.
func (m *Details) NoNotes() bool {
	return m.Notes == ""
}

// NoArtist tests if the photo has no Artist.
func (m *Details) NoArtist() bool {
	return m.Artist == ""
}

// NoCopyright tests if the photo has no Copyright.
func (m *Details) NoCopyright() bool {
	return m.Copyright == ""
}

// NoLicense tests if the photo has no License.
func (m *Details) NoLicense() bool {
	return m.License == ""
}

// NoSoftware tests if the photo has no Software.
func (m *Details) NoSoftware() bool {
	return m.Software == ""
}

// HasKeywords tests if the photo has a Keywords.
func (m *Details) HasKeywords() bool {
	return !m.NoKeywords()
}

// HasSubject tests if the photo has a Subject.
func (m *Details) HasSubject() bool {
	return !m.NoSubject()
}

// HasNotes tests if the photo has a Notes.
func (m *Details) HasNotes() bool {
	return !m.NoNotes()
}

// HasArtist tests if the photo has an Artist.
func (m *Details) HasArtist() bool {
	return !m.NoArtist()
}

// HasCopyright tests if the photo has a Copyright
func (m *Details) HasCopyright() bool {
	return !m.NoCopyright()
}

// HasLicense tests if the photo has a License.
func (m *Details) HasLicense() bool {
	return !m.NoLicense()
}

// HasSoftware tests if the photo has a Software.
func (m *Details) HasSoftware() bool {
	return !m.NoSoftware()
}

// SetKeywords updates the photo details field.
func (m *Details) SetKeywords(data, src string) {
	val := txt.Clip(data, txt.ClipText)

	if val == "" {
		return
	}

	if (SrcPriority[src] < SrcPriority[m.KeywordsSrc]) && m.HasKeywords() {
		// Ignore if priority is lower and keywords already exist.
		return
	}

	if SrcPriority[src] > SrcPriority[m.KeywordsSrc] {
		// Overwrite existing keywords if priority is higher.
		m.Keywords = val
	} else {
		// Merge keywords if priority is the same.
		m.Keywords = txt.MergeWords(m.Keywords, val)
	}

	m.KeywordsSrc = src
}

// SetNotes updates the photo details field.
func (m *Details) SetNotes(data, src string) {
	val := txt.Clip(data, txt.ClipText)

	if val == "" {
		return
	}

	if (SrcPriority[src] < SrcPriority[m.NotesSrc]) && m.HasNotes() {
		return
	}

	m.Notes = val
	m.NotesSrc = src
}

// SetSubject updates the photo details field.
func (m *Details) SetSubject(data, src string) {
	val := txt.Clip(data, txt.ClipShortText)

	if val == "" {
		return
	}

	if (SrcPriority[src] < SrcPriority[m.SubjectSrc]) && m.HasSubject() {
		return
	}

	m.Subject = val
	m.SubjectSrc = src
}

// SetArtist updates the photo details field.
func (m *Details) SetArtist(data, src string) {
	val := txt.Clip(data, txt.ClipShortText)

	if val == "" {
		return
	}

	if (SrcPriority[src] < SrcPriority[m.ArtistSrc]) && m.HasArtist() {
		return
	}

	m.Artist = val
	m.ArtistSrc = src
}

// SetCopyright updates the photo details field.
func (m *Details) SetCopyright(data, src string) {
	val := txt.Clip(data, txt.ClipShortText)

	if val == "" {
		return
	}

	if (SrcPriority[src] < SrcPriority[m.CopyrightSrc]) && m.HasCopyright() {
		return
	}

	m.Copyright = val
	m.CopyrightSrc = src
}

// SetLicense updates the photo details field.
func (m *Details) SetLicense(data, src string) {
	val := txt.Clip(data, txt.ClipShortText)

	if val == "" {
		return
	}

	if (SrcPriority[src] < SrcPriority[m.LicenseSrc]) && m.HasLicense() {
		return
	}

	m.License = val
	m.LicenseSrc = src
}

// SetSoftware updates the photo details field.
func (m *Details) SetSoftware(data, src string) {
	val := txt.Clip(data, txt.ClipShortText)

	if val == "" {
		return
	}

	if (SrcPriority[src] < SrcPriority[m.SoftwareSrc]) && m.HasSoftware() {
		return
	}

	m.Software = val
	m.SoftwareSrc = src
}
