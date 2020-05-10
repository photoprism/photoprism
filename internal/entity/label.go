package entity

import (
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Label is used for photo, album and location categorization
type Label struct {
	ID               uint   `gorm:"primary_key"`
	LabelUUID        string `gorm:"type:varbinary(36);unique_index;"`
	LabelSlug        string `gorm:"type:varbinary(255);unique_index;"`
	CustomSlug       string `gorm:"type:varbinary(255);index;"`
	LabelName        string `gorm:"type:varchar(255);"`
	LabelPriority    int
	LabelFavorite    bool
	LabelDescription string   `gorm:"type:text;"`
	LabelNotes       string   `gorm:"type:text;"`
	LabelCategories  []*Label `gorm:"many2many:categories;association_jointable_foreignkey:category_id"`
	Links            []Link   `gorm:"foreignkey:ShareUUID;association_foreignkey:LabelUUID"`
	PhotoCount       int      `gorm:"default:1"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time `sql:"index"`
	New              bool       `gorm:"-"`
}

// BeforeCreate computes a random UUID when a new label is created in database
func (m *Label) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsPPID(m.LabelUUID, 'l') {
		return nil
	}

	return scope.SetColumn("LabelUUID", rnd.PPID('l'))
}

// NewLabel creates a label in database with a given name and priority
func NewLabel(name string, priority int) *Label {
	labelName := txt.Clip(name, txt.ClipDefault)

	if labelName == "" {
		labelName = "Unknown"
	}

	labelName = txt.Title(labelName)
	labelSlug := slug.Make(txt.Clip(labelName, txt.ClipSlug))

	result := &Label{
		LabelSlug:     labelSlug,
		CustomSlug:    labelSlug,
		LabelName:     labelName,
		LabelPriority: priority,
		PhotoCount:    1,
	}

	return result
}

// FirstOrCreate checks if the label already exists in the database
func (m *Label) FirstOrCreate() *Label {
	if err := Db().FirstOrCreate(m, "label_slug = ? OR custom_slug = ?", m.LabelSlug, m.CustomSlug).Error; err != nil {
		log.Errorf("label: %s", err)
	}

	return m
}

// AfterCreate sets the New column used for database callback
func (m *Label) AfterCreate(scope *gorm.Scope) error {
	return scope.SetColumn("New", true)
}

// SetName changes the label name.
func (m *Label) SetName(name string) {
	newName := txt.Clip(name, txt.ClipDefault)

	if newName == "" {
		return
	}

	m.LabelName = txt.Title(newName)
	m.CustomSlug = slug.Make(txt.Clip(name, txt.ClipSlug))
}

// Updates a label if necessary
func (m *Label) Update(label classify.Label) error {
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

	if save {
		if err := db.Save(m).Error; err != nil {
			return err
		}
	}

	// Add categories
	for _, category := range label.Categories {
		sn := NewLabel(txt.Title(category), -3).FirstOrCreate()
		if err := db.Model(m).Association("LabelCategories").Append(sn).Error; err != nil {
			return err
		}
	}

	return nil
}
