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

var subjectMutex = sync.Mutex{}

// Subjects represents a list of subjects.
type Subjects []Subject

// Subject represents a named photo subject, typically a person.
type Subject struct {
	SubjUID      string          `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"UID" yaml:"UID"`
	MarkerUID    string          `gorm:"type:VARBINARY(42);index" json:"MarkerUID" yaml:"MarkerUID,omitempty"`
	MarkerSrc    string          `gorm:"type:VARBINARY(8);default:''" json:"MarkerSrc,omitempty" yaml:"MarkerSrc,omitempty"`
	SubjType     string          `gorm:"type:VARBINARY(8);default:''" json:"Type,omitempty" yaml:"Type,omitempty"`
	SubjSrc      string          `gorm:"type:VARBINARY(8);default:''" json:"Src,omitempty" yaml:"Src,omitempty"`
	SubjSlug     string          `gorm:"type:VARBINARY(255);index;default:''" json:"Slug" yaml:"-"`
	SubjName     string          `gorm:"type:VARCHAR(255);unique_index;default:''" json:"Name" yaml:"Name"`
	SubjAlias    string          `gorm:"type:VARCHAR(255);default:''" json:"Alias" yaml:"Alias"`
	SubjBio      string          `gorm:"type:TEXT;" json:"Bio" yaml:"Bio,omitempty"`
	SubjNotes    string          `gorm:"type:TEXT;" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	SubjFavorite bool            `gorm:"default:false" json:"Favorite" yaml:"Favorite,omitempty"`
	SubjPrivate  bool            `gorm:"default:false" json:"Private" yaml:"Private,omitempty"`
	SubjExcluded bool            `gorm:"default:false" json:"Excluded" yaml:"Excluded,omitempty"`
	FileCount    int             `gorm:"default:0" json:"FileCount" yaml:"-"`
	MetadataJSON json.RawMessage `gorm:"type:MEDIUMBLOB;" json:"Metadata,omitempty" yaml:"Metadata,omitempty"`
	CreatedAt    time.Time       `json:"CreatedAt" yaml:"-"`
	UpdatedAt    time.Time       `json:"UpdatedAt" yaml:"-"`
	DeletedAt    *time.Time      `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName returns the entity database table name.
func (Subject) TableName() string {
	return "subjects_dev9"
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Subject) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.SubjUID, 'j') {
		return nil
	}

	return scope.SetColumn("SubjUID", rnd.PPID('j'))
}

// NewSubject returns a new entity.
func NewSubject(name, subjType, subjSrc string) *Subject {
	if subjType == "" {
		subjType = SubjPerson
	}

	subjName := txt.Title(txt.Clip(name, txt.ClipDefault))
	subjSlug := slug.Make(txt.Clip(name, txt.ClipSlug))

	// Name is required.
	if subjName == "" || subjSlug == "" {
		return nil
	}

	result := &Subject{
		SubjSlug:  subjSlug,
		SubjName:  subjName,
		SubjType:  subjType,
		SubjSrc:   subjSrc,
		FileCount: 1,
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

// Delete marks the entity as deleted in the database.
func (m *Subject) Delete() error {
	if m.Deleted() {
		return nil
	}

	log.Infof("subject: deleting %s %s", m.SubjType, txt.Quote(m.SubjName))

	event.EntitiesDeleted("subjects", []string{m.SubjUID})

	if m.IsPerson() {
		event.EntitiesDeleted("people", []string{m.SubjUID})
		event.Publish("count.people", event.Data{
			"count": -1,
		})
	}

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

		log.Infof("subject: restoring %s %s", m.SubjType, txt.Quote(m.SubjName))

		event.EntitiesCreated("subjects", []*Subject{m})

		if m.IsPerson() {
			event.EntitiesCreated("people", []*Person{m.Person()})
			event.Publish("count.people", event.Data{
				"count": 1,
			})
		}

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
	} else if m.SubjName == "" {
		return nil
	}

	if found := FindSubjectByName(m.SubjName); found != nil {
		return found
	} else if createErr := m.Create(); createErr == nil {
		log.Infof("subject: added %s %s", m.SubjType, txt.Quote(m.SubjName))

		event.EntitiesCreated("subjects", []*Subject{m})

		if m.IsPerson() {
			event.EntitiesCreated("people", []*Person{m.Person()})
			event.Publish("count.people", event.Data{
				"count": 1,
			})
		}

		return m
	} else if found = FindSubjectByName(m.SubjName); found != nil {
		return found
	} else {
		log.Errorf("subject: %s while creating %s", createErr, txt.Quote(m.SubjName))
	}

	return nil
}

// FindSubject returns an existing entity if exists.
func FindSubject(s string) *Subject {
	if s == "" {
		return nil
	}

	result := Subject{}

	db := Db().Where("subj_uid = ?", s)

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
	db := UnscopedDb().Where("subj_name LIKE ?", s).First(&result)

	if err := db.First(&result).Error; err != nil {
		return nil
	}

	// Restore if currently deleted.
	if err := result.Restore(); err != nil {
		log.Errorf("subject: %s could not be restored", result.SubjUID)
	} else {
		log.Debugf("subject: %s restored", result.SubjUID)
	}

	return &result
}

// IsPerson tests if the subject is a person.
func (m *Subject) IsPerson() bool {
	return m.SubjType == SubjPerson
}

// Person creates and returns a Person based on this subject.
func (m *Subject) Person() *Person {
	return NewPerson(*m)
}

// SetName changes the subject's name.
func (m *Subject) SetName(name string) error {
	newName := txt.Clip(name, txt.ClipDefault)

	if newName == "" {
		return fmt.Errorf("subject: name must not be empty")
	}

	m.SubjName = txt.Title(newName)
	m.SubjSlug = slug.Make(txt.Clip(name, txt.ClipSlug))

	return nil
}

// UpdateName changes and saves the subject's name in the index.
func (m *Subject) UpdateName(name string) (*Subject, error) {
	if err := m.SetName(name); err != nil {
		return m, err
	} else if err := m.Updates(Values{"SubjName": m.SubjName, "SubjSlug": m.SubjSlug}); err == nil {
		log.Infof("subject: renamed %s %s", m.SubjType, txt.Quote(m.SubjName))

		event.EntitiesUpdated("subjects", []*Subject{m})

		if m.IsPerson() {
			event.EntitiesUpdated("people", []*Person{m.Person()})
		}

		return m, m.UpdateMarkerNames()
	} else if existing := FindSubjectByName(m.SubjName); existing == nil {
		return m, err
	} else {
		return existing, m.MergeWith(existing)
	}
}

// UpdateMarkerNames updates related marker names.
func (m *Subject) UpdateMarkerNames() error {
	if m.SubjName == "" {
		return fmt.Errorf("subject name is empty")
	} else if m.SubjUID == "" {
		return fmt.Errorf("subject uid is empty")
	}

	if err := Db().Model(&Marker{}).
		Where("subj_uid = ? AND subj_src <> ?", m.SubjUID, SrcAuto).
		Where("marker_name <> ?", m.SubjName).
		Update(Values{"MarkerName": m.SubjName}).Error; err != nil {
		return err
	}

	return m.RefreshPhotos()
}

// RefreshPhotos flags related photos for metadata maintenance.
func (m *Subject) RefreshPhotos() error {
	if m.SubjUID == "" {
		return fmt.Errorf("empty subject uid")
	}

	return UnscopedDb().Exec(`UPDATE photos SET checked_at = NULL WHERE id IN
		(SELECT f.photo_id FROM files f JOIN ? m ON m.file_uid = f.file_uid WHERE m.subj_uid = ? GROUP BY f.photo_id)`,
		gorm.Expr(Marker{}.TableName()), m.SubjUID).Error
}

// MergeWith merges this subject with another subject and then deletes it.
func (m *Subject) MergeWith(other *Subject) error {
	if other == nil {
		return fmt.Errorf("other subject is nil")
	} else if other.SubjUID == "" {
		return fmt.Errorf("other subject's uid is empty")
	} else if m.SubjUID == "" {
		return fmt.Errorf("subject uid is empty")
	}

	// Update markers and faces with new SubjUID.
	if err := Db().Model(&Marker{}).
		Where("subj_uid = ?", m.SubjUID).
		Update(Values{"SubjUID": other.SubjUID}).Error; err != nil {
		return err
	} else if err := Db().Model(&Face{}).
		Where("subj_uid = ?", m.SubjUID).
		Update(Values{"SubjUID": other.SubjUID}).Error; err != nil {
		return err
	} else if err := other.UpdateMarkerNames(); err != nil {
		return err
	}

	return m.Delete()
}

// Links returns all share links for this entity.
func (m *Subject) Links() Links {
	return FindLinks("", m.SubjUID)
}
