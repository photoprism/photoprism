package models

import (
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

// Camera lens (as extracted from EXIF metadata)
type Lens struct {
	gorm.Model
	LensSlug        string
	LensModel       string
	LensMake        string
	LensType        string
	LensOwner       string
	LensDescription string `gorm:"type:text;"`
	LensNotes       string `gorm:"type:text;"`
}

func (Lens) TableName() string {
	return "lenses"
}

func NewLens(modelName string, makeName string) *Lens {
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

func (c *Lens) FirstOrCreate(db *gorm.DB) *Lens {
	db.FirstOrCreate(c, "lens_model = ? AND lens_make = ?", c.LensModel, c.LensMake)

	return c
}
