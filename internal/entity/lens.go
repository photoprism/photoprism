package entity

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

var lensMutex = sync.Mutex{}

// Lenses represents a list of lenses.
type Lenses []Lens

// Lens represents camera lens (as extracted from UpdateExif metadata)
type Lens struct {
	ID              uint       `gorm:"primary_key" json:"ID" yaml:"ID"`
	LensSlug        string     `gorm:"type:VARBINARY(160);unique_index;" json:"Slug" yaml:"Slug,omitempty"`
	LensName        string     `gorm:"type:VARCHAR(160);" json:"Name" yaml:"Name"`
	LensMake        string     `gorm:"type:VARCHAR(160);" json:"Make" yaml:"Make,omitempty"`
	LensModel       string     `gorm:"type:VARCHAR(160);" json:"Model" yaml:"Model,omitempty"`
	LensType        string     `gorm:"type:VARCHAR(100);" json:"Type" yaml:"Type,omitempty"`
	LensDescription string     `gorm:"type:VARCHAR(2048);" json:"Description,omitempty" yaml:"Description,omitempty"`
	LensNotes       string     `gorm:"type:VARCHAR(1024);" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	CreatedAt       time.Time  `json:"-" yaml:"-"`
	UpdatedAt       time.Time  `json:"-" yaml:"-"`
	DeletedAt       *time.Time `sql:"index" json:"-" yaml:"-"`
}

// TableName returns the entity table name.
func (Lens) TableName() string {
	return "lenses"
}

// UnknownLens is the placeholder used when no lens make or model is known.
var UnknownLens = Lens{
	LensSlug:  UnknownID,
	LensName:  "Unknown",
	LensMake:  "",
	LensModel: "Unknown",
}

// CreateUnknownLens initializes the database with an unknown lens if not exists
func CreateUnknownLens() {
	UnknownLens = *FirstOrCreateLens(&UnknownLens)
}

// NewLens creates a new camera lens entity from make and model names.
func NewLens(makeName string, modelName string) *Lens {
	makeName = strings.TrimSpace(makeName)
	modelName = strings.TrimSpace(modelName)

	if modelName == "" && makeName == "" {
		return &UnknownLens
	} else if strings.HasPrefix(modelName, makeName) {
		modelName = strings.TrimSpace(modelName[len(makeName):])
	}

	// Normalize make name.
	if n, ok := CameraMakes[makeName]; ok {
		makeName = n
	}

	// Remove duplicate make from model name.
	if strings.HasPrefix(modelName, makeName) {
		modelName = strings.TrimSpace(modelName[len(makeName):])
	}

	// Remove ignored substrings from model name.
	modelName = LensModelIgnore.ReplaceAllString(modelName, " ")

	var name []string

	if makeName != "" {
		name = append(name, makeName)
	}

	if modelName != "" {
		name = append(name, modelName)
	}

	lensName := strings.Join(name, " ")

	result := &Lens{
		LensSlug:  txt.Slug(lensName),
		LensName:  txt.Clip(lensName, txt.ClipName),
		LensMake:  txt.Clip(makeName, txt.ClipName),
		LensModel: txt.Clip(modelName, txt.ClipName),
	}

	return result
}

// Create inserts a new row to the database.
func (m *Lens) Create() error {
	lensMutex.Lock()
	defer lensMutex.Unlock()

	return Db().Create(m).Error
}

// FirstOrCreateLens returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateLens(m *Lens) *Lens {
	if m.LensSlug == "" {
		return &UnknownLens
	}

	if cacheData, ok := lensCache.Get(m.LensSlug); ok {
		log.Tracef("lens: cache hit for %s", m.LensSlug)

		return cacheData.(*Lens)
	}

	result := Lens{}

	if res := Db().Where("lens_slug = ?", m.LensSlug).First(&result); res.Error == nil {
		lensCache.SetDefault(m.LensSlug, &result)
		return &result
	} else if err := m.Create(); err == nil {
		if !m.Unknown() {
			// Content channels carry only stable identities, never entity fields; publish the slug.
			event.EntitiesCreated("lenses", []string{m.LensSlug})

			event.Publish("count.lenses", event.Data{
				"count": 1,
			})
		}

		lensCache.SetDefault(m.LensSlug, m)

		return m
	} else if res := Db().Where("lens_slug = ?", m.LensSlug).First(&result); res.Error == nil {
		lensCache.SetDefault(m.LensSlug, &result)
		return &result
	} else {
		log.Errorf("lens: %s (create %s)", err.Error(), clean.Log(m.String()))
	}

	return &UnknownLens
}

// String returns an identifier that can be used in logs.
func (m *Lens) String() string {
	if m == nil {
		return "Lens<nil>"
	}

	return clean.Log(m.LensName)
}

// Unknown returns true if the lens is not a known make or model.
func (m *Lens) Unknown() bool {
	return m.LensSlug == "" || m.LensSlug == UnknownLens.LensSlug
}

// UpdateMakeModel updates the make and model of an existing lens, e.g. to fix Pentax models that
// ExifTool decodes as a numeric "4 38".
// The lens slug is intentionally left unchanged so existing photo references and the unique slug
// index are preserved across renames.
func (m *Lens) UpdateMakeModel(makeName, modelName string) error {
	if m.ID == 0 {
		return fmt.Errorf("empty id")
	}

	makeName = strings.TrimSpace(makeName)
	modelName = strings.TrimSpace(modelName)

	if makeName == "" || modelName == "" {
		return fmt.Errorf("make and model must not be empty")
	}

	l := NewLens(makeName, modelName)
	// Override the changeable fields.
	m.LensMake = l.LensMake
	m.LensModel = l.LensModel
	m.LensName = l.LensName

	lensMutex.Lock()
	defer lensMutex.Unlock()
	if err := Db().Save(m).Error; err != nil {
		return err
	} else {
		if !m.Unknown() {
			event.EntitiesUpdated("lenses", []string{m.LensSlug})
		}
		lensCache.SetDefault(m.LensSlug, m)
	}
	return nil
}

// SaveForm validates the form, copies its data into the lens, and persists it.
func (m *Lens) SaveForm(f *form.Lens) error {
	if f == nil {
		return fmt.Errorf("form is nil")
	} else if err := f.Validate(); err != nil {
		return err
	}

	if err := deepcopier.Copy(m).From(f); err != nil {
		return err
	}

	return m.UpdateMakeModel(f.LensMake, f.LensModel)
}
