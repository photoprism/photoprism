package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Label is used for photo, album and location categorization
type Label struct {
	ID               uint   `gorm:"primary_key"`
	LabelUUID        string `gorm:"type:varbinary(36);unique_index;"`
	LabelSlug        string `gorm:"type:varbinary(128);index;"`
	LabelName        string `gorm:"type:varchar(128);"`
	LabelPriority    int
	LabelFavorite    bool
	LabelDescription string   `gorm:"type:text;"`
	LabelNotes       string   `gorm:"type:text;"`
	LabelCategories  []*Label `gorm:"many2many:categories;association_jointable_foreignkey:category_id"`
	Links            []Link   `gorm:"foreignkey:ShareUUID;association_foreignkey:LabelUUID"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time `sql:"index"`
	New              bool       `gorm:"-"`
}

// BeforeCreate computes a random UUID when a new label is created in database
func (m *Label) BeforeCreate(scope *gorm.Scope) error {
	if err := scope.SetColumn("LabelUUID", rnd.PPID('l')); err != nil {
		log.Errorf("label: %s", err)
		return err
	}

	return nil
}

// NewLabel creates a label in database with a given name and priority
func NewLabel(labelName string, labelPriority int) *Label {
	labelName = strings.TrimSpace(labelName)

	if labelName == "" {
		labelName = "Unknown"
	}

	labelSlug := slug.Make(labelName)

	result := &Label{
		LabelName:     labelName,
		LabelSlug:     labelSlug,
		LabelPriority: labelPriority,
	}

	return result
}

// FirstOrCreate checks wether the label already exists in the database
func (m *Label) FirstOrCreate(db *gorm.DB) *Label {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "label_slug = ?", m.LabelSlug).Error; err != nil {
		log.Errorf("label: %s", err)
	}

	return m
}

// AfterCreate sets the New column used for database callback
func (m *Label) AfterCreate(scope *gorm.Scope) error {
	return scope.SetColumn("New", true)
}

// Rename an existing label
func (m *Label) Rename(name string) {
	name = txt.Clip(name, 128)

	if name == "" {
		name = txt.SlugToTitle(m.LabelSlug)
	}

	m.LabelName = name
}
