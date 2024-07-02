package entity

import (
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/tensorflow/classify"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	LabelUID = byte('l')
)

var labelMutex = sync.Mutex{}
var labelCategoriesMutex = sync.Mutex{}

type Labels []Label

// Label is used for photo, album and location categorization
type Label struct {
	ID               uint       `gorm:"primary_key" json:"ID" yaml:"-"`
	LabelUID         string     `gorm:"type:VARBINARY(42);unique_index;" json:"UID" yaml:"UID"`
	LabelSlug        string     `gorm:"type:VARBINARY(160);unique_index;" json:"Slug" yaml:"-"`
	CustomSlug       string     `gorm:"type:VARBINARY(160);index;" json:"CustomSlug" yaml:"-"`
	LabelName        string     `gorm:"type:VARCHAR(160);" json:"Name" yaml:"Name"`
	LabelPriority    int        `json:"Priority" yaml:"Priority,omitempty"`
	LabelFavorite    bool       `json:"Favorite" yaml:"Favorite,omitempty"`
	LabelDescription string     `gorm:"type:VARCHAR(2048);" json:"Description" yaml:"Description,omitempty"`
	LabelNotes       string     `gorm:"type:VARCHAR(1024);" json:"Notes" yaml:"Notes,omitempty"`
	LabelCategories  []*Label   `gorm:"many2many:categories;association_jointable_foreignkey:category_id" json:"-" yaml:"-"`
	PhotoCount       int        `gorm:"default:1" json:"PhotoCount" yaml:"-"`
	Thumb            string     `gorm:"type:VARBINARY(128);index;default:''" json:"Thumb" yaml:"Thumb,omitempty"`
	ThumbSrc         string     `gorm:"type:VARBINARY(8);default:''" json:"ThumbSrc,omitempty" yaml:"ThumbSrc,omitempty"`
	CreatedAt        time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt        time.Time  `json:"UpdatedAt" yaml:"-"`
	PublishedAt      *time.Time `sql:"index" json:"PublishedAt,omitempty" yaml:"PublishedAt,omitempty"`
	DeletedAt        *time.Time `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
	New              bool       `gorm:"-" json:"-" yaml:"-"`
}

// TableName returns the entity table name.
func (Label) TableName() string {
	return "labels"
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Label) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUnique(m.LabelUID, LabelUID) {
		return nil
	}

	return scope.SetColumn("LabelUID", rnd.GenerateUID(LabelUID))
}

// NewLabel returns a new label.
func NewLabel(name string, priority int) *Label {
	labelName := txt.Clip(name, txt.ClipDefault)

	if labelName == "" {
		labelName = "Unknown"
	}

	labelName = txt.Title(labelName)
	labelSlug := txt.Slug(labelName)

	result := &Label{
		LabelSlug:     labelSlug,
		CustomSlug:    labelSlug,
		LabelName:     txt.Clip(labelName, txt.ClipName),
		LabelPriority: priority,
		PhotoCount:    1,
	}

	return result
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *Label) Save() error {
	labelMutex.Lock()
	defer labelMutex.Unlock()

	return Db().Save(m).Error
}

// Create inserts the label to the database.
func (m *Label) Create() error {
	labelMutex.Lock()
	defer labelMutex.Unlock()

	return Db().Create(m).Error
}

// Delete removes the label from the database.
func (m *Label) Delete() error {
	Db().Where("label_id = ? OR category_id = ?", m.ID, m.ID).Delete(&Category{})
	Db().Where("label_id = ?", m.ID).Delete(&PhotoLabel{})
	return Db().Delete(m).Error
}

// Deleted returns true if the label is deleted.
func (m *Label) Deleted() bool {
	if m.DeletedAt == nil {
		return false
	}

	return !m.DeletedAt.IsZero()
}

// Restore restores the label in the database.
func (m *Label) Restore() error {
	if m.Deleted() {
		return UnscopedDb().Model(m).Update("DeletedAt", nil).Error
	}

	return nil
}

// Update a label property in the database.
func (m *Label) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// FirstOrCreateLabel returns the existing label, inserts a new label or nil in case of errors.
func FirstOrCreateLabel(m *Label) *Label {
	result := Label{}

	if err := UnscopedDb().Where("label_slug = ? OR custom_slug = ?", m.LabelSlug, m.CustomSlug).First(&result).Error; err == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		if m.LabelPriority >= 0 {
			event.EntitiesCreated("labels", []*Label{m})

			event.Publish("count.labels", event.Data{
				"count": 1,
			})
		}

		return m
	} else if err := UnscopedDb().Where("label_slug = ? OR custom_slug = ?", m.LabelSlug, m.CustomSlug).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("label: %s (find or create %s)", createErr, m.LabelSlug)
	}

	return nil
}

// FindLabel returns an existing row if exists.
func FindLabel(s string) *Label {
	labelSlug := txt.Slug(s)

	result := Label{}

	if err := Db().Where("label_slug = ? OR custom_slug = ?", labelSlug, labelSlug).First(&result).Error; err == nil {
		return &result
	}

	return nil
}

// AfterCreate sets the New column used for database callback
func (m *Label) AfterCreate(scope *gorm.Scope) error {
	m.New = true
	return nil
}

// SetName changes the label name.
func (m *Label) SetName(name string) {
	name = clean.NameCapitalized(name)

	if name == "" {
		return
	}

	m.LabelName = txt.Clip(name, txt.ClipName)
	m.CustomSlug = txt.Slug(name)
}

// UpdateClassify updates a label if necessary
func (m *Label) UpdateClassify(label classify.Label) error {
	save := false
	db := Db()

	if m.LabelPriority != label.Priority {
		m.LabelPriority = label.Priority
		save = true
	}

	if m.CustomSlug == "" {
		m.CustomSlug = m.LabelSlug
		save = true
	} else if m.LabelSlug == "" {
		m.LabelSlug = m.CustomSlug
		save = true
	}

	if m.CustomSlug == m.LabelSlug && label.Title() != m.LabelName {
		m.SetName(label.Title())
		save = true
	}

	// Save label.
	if save {
		if err := db.Save(m).Error; err != nil {
			return err
		}
	}

	// Update label categories.
	if len(label.Categories) > 0 {
		labelCategoriesMutex.Lock()
		defer labelCategoriesMutex.Unlock()

		for _, category := range label.Categories {
			sn := FirstOrCreateLabel(NewLabel(txt.Title(category), -3))

			if sn == nil {
				continue
			}

			if sn.Deleted() {
				continue
			}

			if err := db.Model(m).Association("LabelCategories").Append(sn).Error; err != nil {
				log.Debugf("index: failed saving label category %s (%s)", clean.Log(category), err)
			}
		}
	}

	return nil
}

// Links returns all share links for this entity.
func (m *Label) Links() Links {
	return FindLinks("", m.LabelUID)
}
