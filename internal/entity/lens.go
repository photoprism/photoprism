package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/mutex"
)

// Lens reprensent camera lens (as extracted from UpdateExif metadata)
type Lens struct {
	ID              uint   `gorm:"primary_key"`
	LensSlug        string `gorm:"type:varbinary(128);unique_index;"`
	LensModel       string
	LensMake        string
	LensType        string
	LensOwner       string
	LensDescription string `gorm:"type:text;"`
	LensNotes       string `gorm:"type:text;"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time `sql:"index"`
}

// TableName returns Lens table identifier "lens"
func (Lens) TableName() string {
	return "lenses"
}

// NewLens create a new lens in database
func NewLens(modelName string, makeName string) *Lens {
	modelName = strings.TrimSpace(modelName)
	makeName = strings.TrimSpace(makeName)

	if modelName == "" {
		modelName = "Unknown"
	}

	lensSlug := slug.MakeLang(modelName, "en")

	result := &Lens{
		LensModel: modelName,
		LensMake:  makeName,
		LensSlug:  lensSlug,
	}

	return result
}

// FirstOrCreate check wether the lens already exists in the database
func (m *Lens) FirstOrCreate(db *gorm.DB) *Lens {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "lens_slug = ?", m.LensSlug).Error; err != nil {
		log.Errorf("lens: %s", err)
	}

	return m
}
