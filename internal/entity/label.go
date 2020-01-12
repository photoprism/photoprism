package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/rnd"
)

// Labels for photo, album and location categorization
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
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time `sql:"index"`
	New              bool       `gorm:"-"`
}

func (m *Label) BeforeCreate(scope *gorm.Scope) error {
	if err := scope.SetColumn("LabelUUID", rnd.PPID('l')); err != nil {
		log.Errorf("label: %s", err)
		return err
	}

	return nil
}

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

func (m *Label) FirstOrCreate(db *gorm.DB) *Label {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "label_slug = ?", m.LabelSlug).Error; err != nil {
		log.Errorf("label: %s", err)
	}

	return m
}

func (m *Label) AfterCreate(scope *gorm.Scope) error {
	return scope.SetColumn("New", true)
}
