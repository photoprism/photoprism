package entity

import (
	"sync"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

var peopleMutex = sync.Mutex{}

type People []Person

// Person represents a person on one or more photos.
type Person struct {
	ID                uint       `gorm:"primary_key" json:"ID" yaml:"-"`
	PersonUID         string     `gorm:"type:VARBINARY(42);unique_index;" json:"UID" yaml:"UID"`
	PersonSlug        string     `gorm:"type:VARBINARY(255);index;" json:"Slug" yaml:"-"`
	PersonName        string     `gorm:"type:VARCHAR(255);" json:"Name" yaml:"Name"`
	PersonSrc         string     `gorm:"type:VARBINARY(8);" json:"Src" yaml:"Src"`
	PersonFavorite    bool       `json:"Favorite" yaml:"Favorite,omitempty"`
	PersonPrivate     bool       `json:"Private" yaml:"Private,omitempty"`
	PersonHidden      bool       `json:"Hidden" yaml:"Hidden,omitempty"`
	PersonDescription string     `gorm:"type:TEXT;" json:"Description" yaml:"Description,omitempty"`
	PersonNotes       string     `gorm:"type:TEXT;" json:"Notes" yaml:"Notes,omitempty"`
	PersonMeta        string     `gorm:"type:LONGTEXT;" json:"Meta" yaml:"Meta,omitempty"`
	PhotoCount        int        `gorm:"default:0" json:"PhotoCount" yaml:"-"`
	BirthYear         int        `json:"BirthYear" yaml:"BirthYear,omitempty"`
	BirthMonth        int        `json:"BirthMonth" yaml:"BirthMonth,omitempty"`
	BirthDay          int        `json:"BirthDay" yaml:"BirthDay,omitempty"`
	PassedAway        *time.Time `json:"PassedAway" yaml:"PassedAway,omitempty"`
	CreatedAt         time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt         time.Time  `json:"UpdatedAt" yaml:"-"`
	DeletedAt         *time.Time `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// UnknownPerson can be used as a placeholder for unknown people.
var UnknownPerson = Person{
	ID:             1,
	PersonUID:      "r000000000000001",
	PersonSlug:     "zz",
	PersonName:     "Unknown",
	PersonSrc:      SrcDefault,
	PersonFavorite: false,
	BirthYear:      YearUnknown,
	BirthMonth:     MonthUnknown,
	BirthDay:       DayUnknown,
	PhotoCount:     0,
}

// CreateUnknownPerson initializes the database with a placeholder for unknown people if not exists.
func CreateUnknownPerson() {
	FirstOrCreatePerson(&UnknownPerson)
}

// TableName returns the entity database table name.
func (Person) TableName() string {
	return "people_dev2"
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Person) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.PersonUID, 'r') {
		return nil
	}

	return scope.SetColumn("PersonUID", rnd.PPID('r'))
}

// NewPerson returns a new person.
func NewPerson(personName, personSrc string, photoCount int) *Person {
	personName = txt.Title(txt.Clip(personName, txt.ClipDefault))
	personSlug := slug.Make(txt.Clip(personName, txt.ClipSlug))

	result := &Person{
		PersonSlug: personSlug,
		PersonName: personName,
		PersonSrc:  personSrc,
		BirthYear:  YearUnknown,
		BirthMonth: MonthUnknown,
		BirthDay:   DayUnknown,
		PhotoCount: photoCount,
	}

	return result
}

// Save updates the existing or inserts a new person.
func (m *Person) Save() error {
	peopleMutex.Lock()
	defer peopleMutex.Unlock()

	return Db().Save(m).Error
}

// Create inserts the person to the database.
func (m *Person) Create() error {
	peopleMutex.Lock()
	defer peopleMutex.Unlock()

	return Db().Create(m).Error
}

// Delete removes the person from the database.
func (m *Person) Delete() error {
	return Db().Delete(m).Error
}

// Deleted returns true if the person is deleted.
func (m *Person) Deleted() bool {
	return m.DeletedAt != nil
}

// Restore restores the person in the database.
func (m *Person) Restore() error {
	if m.Deleted() {
		return UnscopedDb().Model(m).Update("DeletedAt", nil).Error
	}

	return nil
}

// Update a person property in the database.
func (m *Person) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// FirstOrCreatePerson returns the existing person, inserts a new person or nil in case of errors.
func FirstOrCreatePerson(m *Person) *Person {
	result := Person{}

	if err := UnscopedDb().Where("person_slug = ?", m.PersonSlug).First(&result).Error; err == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		if !m.PersonHidden {
			event.EntitiesCreated("people", []*Person{m})

			event.Publish("count.people", event.Data{
				"count": 1,
			})
		}

		return m
	} else if err := UnscopedDb().Where("person_slug = ?", m.PersonSlug).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("person: %s (find or create %s)", createErr, m.PersonSlug)
	}

	return nil
}

// FindPerson returns an existing row if exists.
func FindPerson(s string) *Person {
	if s == "" {
		return nil
	}

	result := Person{}

	db := Db()

	if rnd.IsPPID(s, 'r') {
		db = db.Where("person_uid = ?", s)
	} else {
		db = db.Where("person_slug = ?", slug.Make(txt.Clip(s, txt.ClipSlug)))
	}

	if err := db.First(&result).Error; err != nil {
		return nil
	}

	return &result
}

// SetName changes the person's name.
func (m *Person) SetName(name string) {
	newName := txt.Clip(name, txt.ClipDefault)

	if newName == "" {
		return
	}

	m.PersonName = txt.Title(newName)
	m.PersonSlug = slug.Make(txt.Clip(name, txt.ClipSlug))
}

// Links returns all share links for this entity.
func (m *Person) Links() Links {
	return FindLinks("", m.PersonUID)
}
