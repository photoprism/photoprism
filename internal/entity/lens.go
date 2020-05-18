package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
)

// Lens represents camera lens (as extracted from UpdateExif metadata)
type Lens struct {
	ID              uint       `gorm:"primary_key" yaml:"LensID"`
	LensSlug        string     `gorm:"type:varbinary(255);unique_index;" yaml:"Slug"`
	LensModel       string     `yaml:"Model"`
	LensMake        string     `yaml:"Make"`
	LensType        string     `yaml:"Type,omitempty"`
	LensDescription string     `gorm:"type:text;" yaml:"Description,omitempty"`
	LensNotes       string     `gorm:"type:text;" yaml:"Notes,omitempty"`
	CreatedAt       time.Time  `yaml:"-"`
	UpdatedAt       time.Time  `yaml:"-"`
	DeletedAt       *time.Time `sql:"index" yaml:"-"`
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
