package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
)

// Lens represents camera lens (as extracted from UpdateExif metadata)
type Lens struct {
	ID              uint   `gorm:"primary_key"`
	LensSlug        string `gorm:"type:varbinary(255);unique_index;"`
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

var UnknownLens = Lens{
	LensModel: "Unknown",
	LensMake:  "",
	LensSlug:  "zz",
}

// CreateUnknownLens initializes the database with an unknown lens if not exists
func CreateUnknownLens() {
	UnknownLens.FirstOrCreate()
}

// TableName returns Lens table identifier "lens"
func (Lens) TableName() string {
	return "lenses"
}

// NewLens creates a new lens in database
func NewLens(modelName string, makeName string) *Lens {
	modelName = strings.TrimSpace(modelName)
	makeName = strings.TrimSpace(makeName)
	lensSlug := slug.MakeLang(modelName, "en")

	if modelName == "" {
		return &UnknownLens
	}

	result := &Lens{
		LensModel: modelName,
		LensMake:  makeName,
		LensSlug:  lensSlug,
	}

	return result
}

// FirstOrCreate checks if the lens already exists in the database
func (m *Lens) FirstOrCreate() *Lens {
	if err := Db().FirstOrCreate(m, "lens_slug = ?", m.LensSlug).Error; err != nil {
		log.Errorf("lens: %s", err)
	}

	return m
}
