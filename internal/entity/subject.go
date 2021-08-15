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

type Subjects []Subject

// Subject represents a named photo subject, typically a person.
type Subject struct {
	ID                 uint            `gorm:"primary_key" json:"ID" yaml:"-"`
	SubjectUID         string          `gorm:"type:VARBINARY(42);unique_index;" json:"UID" yaml:"UID"`
	SubjectType        string          `gorm:"type:VARBINARY(8);" json:"Type" yaml:"Type"`
	SubjectSrc         string          `gorm:"type:VARBINARY(8);" json:"Src" yaml:"Src"`
	SubjectSlug        string          `gorm:"type:VARBINARY(255);index;" json:"Slug" yaml:"-"`
	SubjectName        string          `gorm:"type:VARCHAR(255);index;" json:"Name" yaml:"Name"`
	SubjectDescription string          `gorm:"type:TEXT;" json:"Description" yaml:"Description,omitempty"`
	SubjectNotes       string          `gorm:"type:TEXT;" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	MetadataJSON       json.RawMessage `gorm:"type:MEDIUMBLOB;" json:"Metadata,omitempty" yaml:"Metadata,omitempty"`
	Favorite           bool            `json:"Favorite" yaml:"Favorite,omitempty"`
	Hidden             bool            `json:"Hidden" yaml:"Hidden,omitempty"`
	Private            bool            `json:"Private" yaml:"Private,omitempty"`
	PhotoCount         int             `gorm:"default:0" json:"PhotoCount" yaml:"-"`
	BirthYear          int             `json:"BirthYear" yaml:"BirthYear,omitempty"`
	BirthMonth         int             `json:"BirthMonth" yaml:"BirthMonth,omitempty"`
	BirthDay           int             `json:"BirthDay" yaml:"BirthDay,omitempty"`
	PassedAway         *time.Time      `json:"PassedAway,omitempty" yaml:"PassedAway,omitempty"`
	CreatedAt          time.Time       `json:"CreatedAt" yaml:"-"`
	UpdatedAt          time.Time       `json:"UpdatedAt" yaml:"-"`
	DeletedAt          *time.Time      `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// UnknownPerson can be used as a placeholder for unknown people.
var UnknownPerson = Subject{
	ID:          1,
	SubjectUID:  "j000000000000001",
	SubjectSlug: "unknown",
	SubjectName: "Unknown",
	SubjectType: SubjectPerson,
	SubjectSrc:  SrcDefault,
	Favorite:    false,
	BirthYear:   UnknownYear,
	BirthMonth:  UnknownMonth,
	BirthDay:    UnknownDay,
	PhotoCount:  0,
}

// CreateUnknownPerson initializes the database with a placeholder for unknown people if not exists.
func CreateUnknownPerson() {
	FirstOrCreateSubject(&UnknownPerson)
}

// TableName returns the entity database table name.
func (Subject) TableName() string {
	return "subjects_dev3"
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

	if name == "" {
		name = UnknownName
	}

	subjectName := txt.Title(txt.Clip(name, txt.ClipDefault))
	subjectSlug := slug.Make(txt.Clip(name, txt.ClipSlug))

	result := &Subject{
		SubjectSlug: subjectSlug,
		SubjectName: subjectName,
		SubjectType: subjectType,
		SubjectSrc:  subjectSrc,
		BirthYear:   UnknownYear,
		BirthMonth:  UnknownMonth,
		BirthDay:    UnknownDay,
		PhotoCount:  1,
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

// FirstOrCreateSubject returns the existing subject, inserts a new subject or nil in case of errors.
func FirstOrCreateSubject(m *Subject) *Subject {
	result := Subject{}

	if err := UnscopedDb().Where("subject_type = ? AND subject_slug = ?", m.SubjectType, m.SubjectSlug).First(&result).Error; err == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		if !m.Hidden && m.SubjectType == SubjectPerson {
			event.EntitiesCreated("people", []*Subject{m})

			event.Publish("count.people", event.Data{
				"count": 1,
			})
		}

		return m
	} else if err := UnscopedDb().Where("subject_type = ? AND subject_slug = ?", m.SubjectType, m.SubjectSlug).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("subject: %s (find or create %s)", createErr, m.SubjectSlug)
	}

	return nil
}

// FindSubject returns an existing row if exists.
func FindSubject(s string) *Subject {
	if s == "" {
		return nil
	}

	result := Subject{}

	db := Db()

	if rnd.IsPPID(s, 'j') {
		db = db.Where("subject_uid = ?", s)
	} else {
		db = db.Where("subject_slug = ?", slug.Make(txt.Clip(s, txt.ClipSlug)))
	}

	if err := db.First(&result).Error; err != nil {
		return nil
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
func (m *Subject) UpdateName(name string) error {
	if err := m.SetName(name); err != nil {
		return err
	}

	return m.Updates(Values{"SubjectName": m.SubjectName, "SubjectSlug": m.SubjectSlug})
}

// Links returns all share links for this entity.
func (m *Subject) Links() Links {
	return FindLinks("", m.SubjectUID)
}
