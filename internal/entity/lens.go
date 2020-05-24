package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
)

// Lens represents camera lens (as extracted from UpdateExif metadata)
type Lens struct {
	ID              uint       `gorm:"primary_key" json:"ID" yaml:"ID"`
	LensSlug        string     `gorm:"type:varbinary(255);unique_index;" json:"Slug" yaml:"Slug,omitempty"`
	LensModel       string     `json:"Model" yaml:"Model"`
	LensMake        string     `json:"Make" yaml:"Make"`
	LensType        string     `json:"Type" yaml:"Type,omitempty"`
	LensDescription string     `gorm:"type:text;" json:"Description,omitempty" yaml:"Description,omitempty"`
	LensNotes       string     `gorm:"type:text;" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	CreatedAt       time.Time  `json:"-" yaml:"-"`
	UpdatedAt       time.Time  `json:"-" yaml:"-"`
	DeletedAt       *time.Time `sql:"index" json:"-" yaml:"-"`
}

var UnknownLens = Lens{
	LensModel: "Unknown",
	LensMake:  "",
	LensSlug:  "zz",
}

// CreateUnknownLens initializes the database with an unknown lens if not exists
func CreateUnknownLens() {
	FirstOrCreateLens(&UnknownLens)
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

// Create inserts a new row to the database.
func (m *Lens) Create() error {
	if err := Db().Create(m).Error; err != nil {
		return err
	}

	return nil
}

// FirstOrCreateLens inserts a new row if not exists.
func FirstOrCreateLens(m *Lens) *Lens {
	result := Lens{}

	if err := Db().Where("lens_slug = ?", m.LensSlug).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Errorf("lens: %s", err)
		return nil
	}

	return m
}
