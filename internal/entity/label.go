package entity

import (
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/event"

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
	ID               uint           `gorm:"primaryKey;" json:"ID" yaml:"-"`
	LabelUID         string         `gorm:"type:bytes;size:42;uniqueIndex;" json:"UID" yaml:"UID"`
	LabelSlug        string         `gorm:"type:bytes;size:160;uniqueIndex;" json:"Slug" yaml:"-"`
	CustomSlug       string         `gorm:"type:bytes;size:160;index;" json:"CustomSlug" yaml:"-"`
	LabelName        string         `gorm:"size:160;" json:"Name" yaml:"Name"`
	LabelPriority    int            `json:"Priority" yaml:"Priority,omitempty"`
	LabelFavorite    bool           `json:"Favorite" yaml:"Favorite,omitempty"`
	LabelDescription string         `gorm:"size:2048;" json:"Description" yaml:"Description,omitempty"`
	LabelNotes       string         `gorm:"size:1024;" json:"Notes" yaml:"Notes,omitempty"`
	LabelCategories  []*Label       `gorm:"many2many:categories;foreignKey:ID;joinForeignKey:LabelID;References:ID;joinReferences:CategoryID" json:"-" yaml:"-"`
	PhotoCount       int            `gorm:"default:1" json:"PhotoCount" yaml:"-"`
	Thumb            string         `gorm:"type:bytes;size:128;index;default:''" json:"Thumb" yaml:"Thumb,omitempty"`
	ThumbSrc         string         `gorm:"type:bytes;size:8;default:''" json:"ThumbSrc,omitempty" yaml:"ThumbSrc,omitempty"`
	CreatedAt        time.Time      `json:"CreatedAt" yaml:"-"`
	UpdatedAt        time.Time      `json:"UpdatedAt" yaml:"-"`
	PublishedAt      *time.Time     `sql:"index" json:"PublishedAt,omitempty" yaml:"PublishedAt,omitempty"`
	DeletedAt        gorm.DeletedAt `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
	New              bool           `gorm:"-" json:"-" yaml:"-"`
}

// TableName returns the entity table name.
func (Label) TableName() string {
	return "labels"
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Label) BeforeCreate(scope *gorm.DB) error {
	if rnd.IsUnique(m.LabelUID, LabelUID) {
		return nil
	}

	scope.Statement.SetColumn("LabelUID", rnd.GenerateUID(LabelUID))
	return scope.Error
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
	return m.DeletedAt.Valid
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
func (m *Label) AfterCreate(scope *gorm.DB) error {
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

			if err := db.Model(m).Association("LabelCategories").Append(sn); err != nil {
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
