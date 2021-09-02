package entity

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	SubjectPerson = "person"
)

var subjectMutex = sync.Mutex{}

// Subjects represents a list of subjects.
type Subjects []Subject

// Subject represents a named photo subject, typically a person.
type Subject struct {
	SubjectUID   string          `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"UID" yaml:"UID"`
	Thumb        string          `gorm:"type:VARBINARY(128);index;default:''" json:"Thumb,omitempty" yaml:"Thumb,omitempty"`
	ThumbSrc     string          `gorm:"type:VARBINARY(8);default:''" json:"ThumbSrc,omitempty" yaml:"ThumbSrc,omitempty"`
	SubjectType  string          `gorm:"type:VARBINARY(8);default:''" json:"Type,omitempty" yaml:"Type,omitempty"`
	SubjectSrc   string          `gorm:"type:VARBINARY(8);default:''" json:"Src,omitempty" yaml:"Src,omitempty"`
	SubjectSlug  string          `gorm:"type:VARBINARY(255);index;default:''" json:"Slug" yaml:"-"`
	SubjectName  string          `gorm:"type:VARCHAR(255);unique_index;default:''" json:"Name" yaml:"Name"`
	SubjectAlias string          `gorm:"type:VARCHAR(255);default:''" json:"Alias" yaml:"Alias"`
	SubjectBio   string          `gorm:"type:TEXT;default:''" json:"Bio" yaml:"Bio,omitempty"`
	SubjectNotes string          `gorm:"type:TEXT;default:''" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	Favorite     bool            `gorm:"default:false" json:"Favorite" yaml:"Favorite,omitempty"`
	Private      bool            `gorm:"default:false" json:"Private" yaml:"Private,omitempty"`
	Excluded     bool            `gorm:"default:false" json:"Excluded" yaml:"Excluded,omitempty"`
	FileCount    int             `gorm:"default:0" json:"Files" yaml:"-"`
	MetadataJSON json.RawMessage `gorm:"type:MEDIUMBLOB;" json:"Metadata,omitempty" yaml:"Metadata,omitempty"`
	CreatedAt    time.Time       `json:"CreatedAt" yaml:"-"`
	UpdatedAt    time.Time       `json:"UpdatedAt" yaml:"-"`
	DeletedAt    *time.Time      `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName returns the entity database table name.
func (Subject) TableName() string {
	return "subjects_dev6"
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Subject) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.SubjectUID, 'j') {
		return nil
	}

	return scope.SetColumn("SubjectUID", rnd.PPID('j'))
}

// NewSubject returns a new entity.
func NewSubject(name, subjectType, subjectSrc string) *Subject {
	if subjectType == "" {
		subjectType = SubjectPerson
	}

	subjectName := txt.Title(txt.Clip(name, txt.ClipDefault))
	subjectSlug := slug.Make(txt.Clip(name, txt.ClipSlug))

	// Name is required.
	if subjectName == "" || subjectSlug == "" {
		return nil
	}

	result := &Subject{
		SubjectSlug: subjectSlug,
		SubjectName: subjectName,
		SubjectType: subjectType,
		SubjectSrc:  subjectSrc,
		FileCount:   1,
	}

	return result
}

// Save updates the existing or inserts a new entity.
func (m *Subject) Save() error {
	subjectMutex.Lock()
	defer subjectMutex.Unlock()

	return Db().Save(m).Error
}

// Create inserts the entity to the database.
func (m *Subject) Create() error {
	subjectMutex.Lock()
	defer subjectMutex.Unlock()

	return Db().Create(m).Error
}

// Delete removes the entity from the database.
func (m *Subject) Delete() error {
	return Db().Delete(m).Error
}

// Deleted returns true if the entity is deleted.
func (m *Subject) Deleted() bool {
	return m.DeletedAt != nil
}

// Restore restores the entity in the database.
func (m *Subject) Restore() error {
	if m.Deleted() {
		m.DeletedAt = nil
		return UnscopedDb().Model(m).Update("DeletedAt", nil).Error
	}

	return nil
}

// Update updates an entity value in the database.
func (m *Subject) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Updates multiple values in the database.
func (m *Subject) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// FirstOrCreateSubject returns the existing entity, inserts a new entity or nil in case of errors.
func FirstOrCreateSubject(m *Subject) *Subject {
	if m == nil {
		return nil
	} else if m.SubjectName == "" {
		return nil
	}

	if found := FindSubjectByName(m.SubjectName); found != nil {
		return found
	} else if createErr := m.Create(); createErr == nil {
		if !m.Excluded && m.SubjectType == SubjectPerson {
			event.EntitiesCreated("people", []*Subject{m})

			event.Publish("count.people", event.Data{
				"count": 1,
			})
		}

		return m
	} else if found = FindSubjectByName(m.SubjectName); found != nil {
		return found
	} else {
		log.Errorf("subject: %s while creating %s", createErr, txt.Quote(m.SubjectName))
	}

	return nil
}

// FindSubject returns an existing entity if exists.
func FindSubject(s string) *Subject {
	if s == "" {
		return nil
	}

	result := Subject{}

	db := Db().Where("subject_uid = ?", s)

	if err := db.First(&result).Error; err != nil {
		return nil
	}

	return &result
}

// FindSubjectByName find an existing subject by name.
func FindSubjectByName(s string) *Subject {
	if s == "" {
		return nil
	}

	result := Subject{}

	// Search database.
	db := UnscopedDb().Where("subject_name LIKE ?", s).First(&result)

	if err := db.First(&result).Error; err != nil {
		return nil
	}

	// Restore if currently deleted.
	if err := result.Restore(); err != nil {
		log.Errorf("subject: %s could not be restored", result.SubjectUID)
	} else {
		log.Debugf("subject: %s restored", result.SubjectUID)
	}

	return &result
}

// SetName changes the subject's name.
func (m *Subject) SetName(name string) error {
	newName := txt.Clip(name, txt.ClipDefault)

	if newName == "" {
		return fmt.Errorf("subject: name must not be empty")
	}

	m.SubjectName = txt.Title(newName)
	m.SubjectSlug = slug.Make(txt.Clip(name, txt.ClipSlug))

	return nil
}

// UpdateName changes and saves the subject's name in the index.
func (m *Subject) UpdateName(name string) (*Subject, error) {
	if err := m.SetName(name); err != nil {
		return m, err
	} else if err := m.Updates(Values{"SubjectName": m.SubjectName, "SubjectSlug": m.SubjectSlug}); err == nil {
		return m, m.UpdateMarkerNames()
	} else if existing := FindSubjectByName(m.SubjectName); existing == nil {
		return m, err
	} else {
		return existing, m.MergeWith(existing)
	}
}

// UpdateMarkerNames updates related marker names.
func (m *Subject) UpdateMarkerNames() error {
	return Db().Model(&Marker{}).
		Where("subject_uid = ? AND subject_src <> ?", m.SubjectUID, SrcAuto).
		Where("marker_name <> '' AND marker_name <> ?", m.SubjectName).
		Update(Values{"MarkerName": m.SubjectName}).Error
}

// MergeWith merges this subject with another subject and then deletes it.
func (m *Subject) MergeWith(other *Subject) error {
	if other == nil {
		return fmt.Errorf("other subject is nil")
	} else if other.SubjectUID == "" {
		return fmt.Errorf("other subject's uid is empty")
	} else if m.SubjectUID == "" {
		return fmt.Errorf("subject uid is empty")
	}

	// Update markers and faces with new SubjectUID.
	if err := Db().Model(&Marker{}).
		Where("subject_uid = ?", m.SubjectUID).
		Update(Values{"SubjectUID": other.SubjectUID}).Error; err != nil {
		return err
	} else if err := Db().Model(&Face{}).
		Where("subject_uid = ?", m.SubjectUID).
		Update(Values{"SubjectUID": other.SubjectUID}).Error; err != nil {
		return err
	} else if err := other.UpdateMarkerNames(); err != nil {
		return err
	}

	return m.Delete()
}

// Links returns all share links for this entity.
func (m *Subject) Links() Links {
	return FindLinks("", m.SubjectUID)
}
