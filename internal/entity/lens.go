package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Lens represents camera lens (as extracted from UpdateExif metadata)
type Lens struct {
	ID              uint       `gorm:"primary_key" json:"ID" yaml:"ID"`
	LensSlug        string     `gorm:"type:varbinary(255);unique_index;" json:"Slug" yaml:"Slug,omitempty"`
	LensName        string     `gorm:"type:varchar(255);" json:"Name" yaml:"Name"`
	LensMake        string     `json:"Make" yaml:"Make,omitempty"`
	LensModel       string     `json:"Model" yaml:"Model,omitempty"`
	LensType        string     `json:"Type" yaml:"Type,omitempty"`
	LensDescription string     `gorm:"type:text;" json:"Description,omitempty" yaml:"Description,omitempty"`
	LensNotes       string     `gorm:"type:text;" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	CreatedAt       time.Time  `json:"-" yaml:"-"`
	UpdatedAt       time.Time  `json:"-" yaml:"-"`
	DeletedAt       *time.Time `sql:"index" json:"-" yaml:"-"`
}

var UnknownLens = Lens{
	LensSlug:  "zz",
	LensName:  "Unknown",
	LensMake:  "",
	LensModel: "Unknown",
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
	modelName = txt.Clip(modelName, txt.ClipDefault)
	makeName = txt.Clip(makeName, txt.ClipDefault)

	if modelName == "" && makeName == "" {
		return &UnknownLens
	} else if strings.HasPrefix(modelName, makeName) {
		modelName = strings.TrimSpace(modelName[len(makeName):])
	}

	if n, ok := CameraMakes[makeName]; ok {
		makeName = n
	}

	var name []string

	if makeName != "" {
		name = append(name, makeName)
	}

	if modelName != "" {
		name = append(name, modelName)
	}

	lensName := strings.Join(name, " ")
	lensSlug := slug.Make(txt.Clip(lensName, txt.ClipSlug))

	result := &Lens{
		LensSlug:  lensSlug,
		LensName:  lensName,
		LensMake:  makeName,
		LensModel: modelName,
	}

	return result
}

// Create inserts a new row to the database.
func (m *Lens) Create() error {
	return Db().Create(m).Error
}

// FirstOrCreateLens returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateLens(m *Lens) *Lens {
	result := Lens{}

	if err := Db().Where("lens_slug = ?", m.LensSlug).First(&result).Error; err == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		if !m.Unknown() {
			event.EntitiesCreated("lenses", []*Lens{m})

			event.Publish("count.lenses", event.Data{
				"count": 1,
			})
		}

		return m
	} else if err := Db().Where("lens_slug = ?", m.LensSlug).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("lens: %s (first or create %s)", createErr, m.String())
	}

	return nil
}

// String returns an identifier that can be used in logs.
func (m *Lens) String() string {
	return m.LensName
}

// Unknown returns true if the lens is not a known make or model.
func (m *Lens) Unknown() bool {
	return m.LensSlug == "" || m.LensSlug == UnknownLens.LensSlug
}
