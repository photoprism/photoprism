package entity

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

var subjectMutex = sync.Mutex{}

// Subject represents a named photo subject, typically a person.
type Subject struct {
	SubjUID      string     `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"UID" yaml:"UID"`
	SubjType     string     `gorm:"type:VARBINARY(8);default:'';" json:"Type,omitempty" yaml:"Type,omitempty"`
	SubjSrc      string     `gorm:"type:VARBINARY(8);default:'';" json:"Src,omitempty" yaml:"Src,omitempty"`
	SubjSlug     string     `gorm:"type:VARBINARY(160);index;default:'';" json:"Slug" yaml:"-"`
	SubjName     string     `gorm:"type:VARCHAR(160);unique_index;default:'';" json:"Name" yaml:"Name"`
	SubjAlias    string     `gorm:"type:VARCHAR(160);default:'';" json:"Alias" yaml:"Alias"`
	SubjBio      string     `gorm:"type:VARCHAR(2048);" json:"Bio" yaml:"Bio,omitempty"`
	SubjNotes    string     `gorm:"type:VARCHAR(1024);" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	SubjFavorite bool       `gorm:"default:false;" json:"Favorite" yaml:"Favorite,omitempty"`
	SubjHidden   bool       `gorm:"default:false;" json:"Hidden" yaml:"Hidden,omitempty"`
	SubjPrivate  bool       `gorm:"default:false;" json:"Private" yaml:"Private,omitempty"`
	SubjExcluded bool       `gorm:"default:false;" json:"Excluded" yaml:"Excluded,omitempty"`
	FileCount    int        `gorm:"default:0;" json:"FileCount" yaml:"-"`
	PhotoCount   int        `gorm:"default:0;" json:"PhotoCount" yaml:"-"`
	Thumb        string     `gorm:"type:VARBINARY(128);index;default:'';" json:"Thumb" yaml:"Thumb,omitempty"`
	ThumbSrc     string     `gorm:"type:VARBINARY(8);default:'';" json:"ThumbSrc,omitempty" yaml:"ThumbSrc,omitempty"`
	CreatedAt    time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt    time.Time  `json:"UpdatedAt" yaml:"-"`
	DeletedAt    *time.Time `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName returns the entity table name.
func (Subject) TableName() string {
	return "subjects"
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Subject) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUnique(m.SubjUID, 'j') {
		return nil
	}

	return scope.SetColumn("SubjUID", rnd.GenerateUID('j'))
}

// AfterSave is a hook that updates the name cache after saving.
func (m *Subject) AfterSave() (err error) {
	SubjNames.Set(m.SubjUID, m.SubjName)
	return
}

// AfterFind is a hook that updates the name cache after querying.
func (m *Subject) AfterFind() (err error) {
	SubjNames.Set(m.SubjUID, m.SubjName)
	return
}

// NewSubject returns a new entity.
func NewSubject(name, subjType, subjSrc string) *Subject {
	// Name is required.
	if strings.TrimSpace(name) == "" {
		return nil
	}

	if subjType == "" {
		subjType = SubjPerson
	}

	result := &Subject{
		SubjType:  subjType,
		SubjSrc:   subjSrc,
		FileCount: 1,
	}

	if err := result.SetName(name); err != nil {
		log.Errorf("subject: %s", err)
	}

	return result
}

// Save updates the record in the database or inserts a new record if it does not already exist.
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

	subjectMutex.Lock()
	defer subjectMutex.Unlock()

	event.EntitiesDeleted("subjects", []string{m.SubjUID})

	if m.IsPerson() {
		event.EntitiesDeleted("people", []string{m.SubjUID})
		event.Publish("count.people", event.Data{
			"count": -1,
		})
	}

	if err := Db().Model(&Face{}).Where("subj_uid = ?", m.SubjUID).Update("subj_uid", "").Error; err != nil {
		return err
	}

	log.Infof("subject: flagged %s as missing", TypeString(m.SubjType), clean.Log(m.SubjName))

	return Db().Delete(m).Error
}

// AfterDelete resets file and photo counters when the entity was deleted.
func (m *Subject) AfterDelete(tx *gorm.DB) (err error) {
	tx.Model(m).Updates(Values{
		"FileCount":  0,
		"PhotoCount": 0,
	})
	return
}

// Deleted returns true if the entity is deleted.
func (m *Subject) Deleted() bool {
	if m.DeletedAt == nil {
		return false
	}

	return !m.DeletedAt.IsZero()
}

// Restore restores the entity in the database.
func (m *Subject) Restore() error {
	if m.Deleted() {
		m.DeletedAt = nil

		log.Infof("subject: restoring %s %s", TypeString(m.SubjType), clean.Log(m.SubjName))

		event.EntitiesCreated("subjects", []*Subject{m})

		if m.IsPerson() {
			event.EntitiesCreated("people", []*Person{m.Person()})
			event.Publish("count.people", event.Data{
				"count": 1,
			})
		}

		return UnscopedDb().Model(m).UpdateColumn("DeletedAt", nil).Error
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
	} else if err := m.Create(); err == nil {
		log.Infof("subject: added %s %s", TypeString(m.SubjType), clean.Log(m.SubjName))

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
		log.Errorf("subject: failed adding %s (%s)", clean.Log(m.SubjName), err)
	}

	return nil
}

// FindSubject returns an existing entity if exists.
func FindSubject(uid string) *Subject {
	if uid == "" {
		return nil
	}

	result := Subject{}

	if err := UnscopedDb().Where("subj_uid = ?", uid).First(&result).Error; err != nil {
		return nil
	}

	return &result
}

// FindSubjectByName find an existing subject by name.
func FindSubjectByName(name string) *Subject {
	name = clean.Name(name)

	if name == "" {
		return nil
	}

	result := Subject{}
	uid := SubjNames.Key(name)

	switch uid {
	case "":
		if err := UnscopedDb().Where("subj_name LIKE ?", name).First(&result).Error; err != nil {
			log.Debugf("subject: %s not found by name", clean.Log(name))
			return nil
		}
	default:
		if found := FindSubject(uid); found == nil {
			log.Debugf("subject: %s not found by uid", clean.Log(name))
			return nil
		} else {
			result = *found
		}
	}

	// Restore if flagged as deleted.
	if !result.Deleted() {
		return &result
	} else if err := result.Restore(); err == nil {
		log.Debugf("subject: restored %s", clean.Log(result.SubjName))
		return &result
	} else {
		log.Errorf("subject: failed restoring %s (%s)", clean.Log(result.SubjName), err)
	}

	return nil
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
	name = clean.Name(name)

	if name == m.SubjName {
		// Nothing to do.
		return nil
	} else if name == "" {
		return fmt.Errorf("name must not be empty")
	}

	m.SubjName = name
	m.SubjSlug = txt.Slug(name)

	return nil
}

// Visible tests if the subject is generally visible and not hidden in any way.
func (m *Subject) Visible() bool {
	return m.DeletedAt == nil && !m.SubjHidden && !m.SubjExcluded && !m.SubjPrivate
}

// SaveForm updates the subject from form values.
func (m *Subject) SaveForm(f form.Subject) (changed bool, err error) {
	if m.SubjUID == "" {
		return false, fmt.Errorf("subject uid is empty")
	}

	// Change name?
	if name := clean.Name(f.SubjName); name != "" && name != m.SubjName {
		existing, err := m.UpdateName(name)

		if existing.SubjUID != m.SubjUID || err != nil {
			return err != nil, err
		}

		changed = true
	}

	// Change favorite status?
	if m.SubjFavorite != f.SubjFavorite {
		m.SubjFavorite = f.SubjFavorite
		changed = true
	}

	// Change visibility?
	if m.SubjHidden != f.SubjHidden || m.SubjPrivate != f.SubjPrivate || m.SubjExcluded != f.SubjExcluded {
		m.SubjHidden = f.SubjHidden
		m.SubjPrivate = f.SubjPrivate
		m.SubjExcluded = f.SubjExcluded

		// Update counter.
		if !m.IsPerson() {
			// Ignore.
		} else if m.Visible() {
			event.Publish("count.people", event.Data{
				"count": 1,
			})
		} else {
			event.Publish("count.people", event.Data{
				"count": -1,
			})
		}

		changed = true
	}

	// Update index?
	if changed {
		values := Values{
			"SubjFavorite": m.SubjFavorite,
			"SubjHidden":   m.SubjHidden,
			"SubjPrivate":  m.SubjPrivate,
			"SubjExcluded": m.SubjExcluded,
		}

		if err := m.Updates(values); err == nil {
			event.EntitiesUpdated("subjects", []*Subject{m})

			if m.IsPerson() {
				event.EntitiesUpdated("people", []*Person{m.Person()})
			}

			return true, nil
		} else {
			return false, err
		}
	}

	return false, nil
}

// UpdateName changes and saves the subject's name in the index.
func (m *Subject) UpdateName(name string) (*Subject, error) {
	if err := m.SetName(name); err != nil {
		return m, err
	} else if err := m.Updates(Values{"SubjName": m.SubjName, "SubjSlug": m.SubjSlug}); err == nil {
		log.Infof("subject: renamed %s %s", TypeString(m.SubjType), clean.Log(m.SubjName))

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
		UpdateColumn("marker_name", m.SubjName).Error; err != nil {
		return err
	}

	return m.RefreshPhotos()
}

// RefreshPhotos flags related photos for metadata maintenance.
func (m *Subject) RefreshPhotos() error {
	if m.SubjUID == "" {
		return fmt.Errorf("empty subject uid")
	}

	var err error
	switch DbDialect() {
	case MySQL:
		update := fmt.Sprintf(`UPDATE photos p JOIN files f ON f.photo_id = p.id JOIN %s m ON m.file_uid = f.file_uid
			SET p.checked_at = NULL WHERE m.subj_uid = ?`, Marker{}.TableName())
		err = UnscopedDb().Exec(update, m.SubjUID).Error
	default:
		update := fmt.Sprintf(`UPDATE photos SET checked_at = NULL WHERE id IN (SELECT f.photo_id FROM files f
			JOIN %s m ON m.file_uid = f.file_uid WHERE m.subj_uid = ?)`, Marker{}.TableName())
		err = UnscopedDb().Exec(update, m.SubjUID).Error
	}

	return err
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
		UpdateColumn("subj_uid", other.SubjUID).Error; err != nil {
		return err
	} else if err := Db().Model(&Face{}).
		Where("subj_uid = ?", m.SubjUID).
		UpdateColumn("subj_uid", other.SubjUID).Error; err != nil {
		return err
	} else if err := other.UpdateMarkerNames(); err != nil {
		return err
	}

	// Update file and photo counts.
	if err := Db().Model(other).Updates(Values{
		"FileCount":  other.FileCount + m.FileCount,
		"PhotoCount": other.PhotoCount + m.PhotoCount,
	}).Error; err != nil {
		return err
	}

	return m.Delete()
}

// Links returns all share links for this entity.
func (m *Subject) Links() Links {
	return FindLinks("", m.SubjUID)
}
