package models

import (
	"strings"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

// Labels for photo, album and location categorization
type Label struct {
	Model
	LabelSlug        string `gorm:"type:varchar(128);index;"`
	LabelName        string `gorm:"type:varchar(128);"`
	LabelPriority    int
	LabelFavorite    bool
	LabelDescription string   `gorm:"type:text;"`
	LabelNotes       string   `gorm:"type:text;"`
	LabelCategories  []*Label `gorm:"many2many:categories;association_jointable_foreignkey:category_id"`
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
	db.FirstOrCreate(m, "label_slug = ?", m.LabelSlug)

	return m
}
