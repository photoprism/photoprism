package entity

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
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
	LensSrc         string     `gorm:"type:VARBINARY(8);default:'';" json:"-" yaml:"-"`
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
	}

	modelName = trimMakePrefix(modelName, makeName)

	// Normalize make name.
	if n, ok := CameraMakes[makeName]; ok {
		makeName = n
	}

	// Remove duplicate make from model name.
	modelName = trimMakePrefix(modelName, makeName)

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

// AddLens returns the lens with the specified make and model, and creates it if it does not exist yet.
// The lens is marked as added manually, so that purging orphans keeps it while no picture references it.
func AddLens(makeName, modelName string) (result *Lens, created bool, err error) {
	makeName = strings.TrimSpace(makeName)
	modelName = strings.TrimSpace(modelName)

	if makeName == "" || modelName == "" {
		return nil, false, fmt.Errorf("%w: make and model must not be empty", ErrInvalidValue)
	}

	m := NewLens(makeName, modelName)

	if strings.TrimSpace(m.LensModel) == "" {
		return nil, false, fmt.Errorf("%w: model must not be empty after removing the make", ErrInvalidValue)
	} else if m.Unknown() {
		return nil, false, fmt.Errorf("%w: make and model must not match the unknown lens", ErrInvalidValue)
	}

	// Report an existing lens instead of creating a duplicate, unless it has been purged in the meantime.
	if existing := lookupExistingLens(m); existing != nil && !existing.Unknown() {
		if err = existing.markManual(); err == nil {
			return existing, false, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, err
		}
	}

	m.LensSrc = SrcManual

	if result = FirstOrCreateLens(m); result == m {
		return result, true, nil
	} else if result.Unknown() {
		return nil, false, fmt.Errorf("failed to create lens %s", m.String())
	} else if err = result.markManual(); err != nil {
		return nil, false, err
	}

	return result, false, nil
}

// lookupExistingLens finds an existing lens when adding one, and can be replaced in tests to simulate a concurrent purge.
var lookupExistingLens = findExistingLens

// findExistingLens returns the lens with the same slug, or else with the same make and model, if any.
// The make and model lookup covers renamed records, whose slug no longer matches their name.
func findExistingLens(m *Lens) *Lens {
	existing := Lens{}

	if Db().Where("lens_slug = ?", m.LensSlug).First(&existing).Error == nil {
		return &existing
	}

	if found := findLensesByMakeModel(m.LensMake, m.LensModel); len(found) > 0 {
		return &found[0]
	}

	return nil
}

// FindLensesByMakeModel returns the lenses that match the make and model as stored, or else after normalizing them.
// It ignores the slug, which a renamed record keeps and may therefore belong to another name.
func FindLensesByMakeModel(makeName, modelName string) Lenses {
	makeName = strings.TrimSpace(makeName)
	modelName = strings.TrimSpace(modelName)

	if makeName == "" || modelName == "" {
		return nil
	} else if found := findLensesByMakeModel(makeName, modelName); len(found) > 0 {
		// Records saved with a different normalization, e.g. by an older version, match as stored.
		return found
	}

	m := NewLens(makeName, modelName)

	if m.Unknown() || strings.TrimSpace(m.LensModel) == "" {
		return nil
	}

	return findLensesByMakeModel(m.LensMake, m.LensModel)
}

// findLensesByMakeModel returns the lenses with exactly the specified make and model, ordered by ID.
func findLensesByMakeModel(makeName, modelName string) (result Lenses) {
	var candidates Lenses

	if Db().Where("lens_make = ? AND lens_model = ?", makeName, modelName).Order("id").Find(&candidates).Error != nil {
		return nil
	}

	// Compare in Go, since MariaDB collations also match values that differ in case, accents, or emoji.
	for i := range candidates {
		if candidates[i].LensMake == makeName && candidates[i].LensModel == modelName && !candidates[i].Unknown() {
			result = append(result, candidates[i])
		}
	}

	return result
}

// markManual marks an existing lens as added manually, so that purging orphans keeps it.
// It returns gorm.ErrRecordNotFound if the lens no longer exists, e.g. because it has been purged.
func (m *Lens) markManual() error {
	if m.ID == 0 {
		// Without a primary key, the update would apply to all rows.
		return fmt.Errorf("empty id")
	}

	lensMutex.Lock()
	defer lensMutex.Unlock()

	if res := UnscopedDb().Model(m).UpdateColumn("lens_src", SrcManual); res.Error != nil {
		return res.Error
	} else if res.RowsAffected == 0 {
		// MariaDB counts only changed rows, so check whether the lens still exists.
		var count int

		if err := UnscopedDb().Model(&Lens{}).Where("id = ?", m.ID).Count(&count).Error; err != nil {
			return err
		} else if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}

	m.LensSrc = SrcManual
	lensCache.SetDefault(m.LensSlug, m)

	return nil
}

// unknownLensID returns the ID of the unknown lens placeholder, and initializes the placeholder if needed.
func unknownLensID() (uint, error) {
	if UnknownLens.ID == 0 {
		CreateUnknownLens()
	}

	if UnknownLens.ID == 0 {
		return 0, fmt.Errorf("unknown lens not found")
	}

	return UnknownLens.ID, nil
}

// PhotoCount returns the number of pictures that reference the lens, including archived and deleted pictures.
func (m *Lens) PhotoCount() (count int, err error) {
	if m.ID == 0 {
		return 0, fmt.Errorf("empty id")
	}

	err = UnscopedDb().Model(&Photo{}).Where("lens_id = ?", m.ID).Count(&count).Error

	return count, err
}

// Delete permanently removes the lens, and first assigns the pictures that reference it to the unknown lens if requested.
// It returns ErrInUse if pictures still reference the lens, so that none of them is left with a dangling reference.
func (m *Lens) Delete(reassign bool) (reassigned int64, err error) {
	if m.ID == 0 {
		return 0, fmt.Errorf("empty id")
	}

	// Resolve the placeholder from the database, since the CLI does not initialize it on startup.
	unknownID, err := unknownLensID()

	if err != nil {
		return 0, err
	} else if m.Unknown() || m.ID == unknownID {
		return 0, fmt.Errorf("%w: unknown lens cannot be deleted", ErrInvalidValue)
	}

	lensMutex.Lock()
	defer lensMutex.Unlock()

	if reassign {
		res := UnscopedDb().Model(&Photo{}).Where("lens_id = ?", m.ID).UpdateColumn("lens_id", unknownID)

		if res.Error != nil {
			return 0, res.Error
		}

		reassigned = res.RowsAffected
	}

	// Delete the lens only if no picture references it, e.g. after being assigned to it by a concurrent index run.
	res := UnscopedDb().Exec(`DELETE FROM lenses WHERE id = ? AND NOT EXISTS (SELECT 1 FROM photos WHERE photos.lens_id = lenses.id)`, m.ID)

	if res.Error != nil {
		return reassigned, res.Error
	} else if res.RowsAffected == 0 {
		var count int

		if err = UnscopedDb().Model(&Lens{}).Where("id = ?", m.ID).Count(&count).Error; err != nil {
			return reassigned, err
		} else if count == 0 {
			return reassigned, gorm.ErrRecordNotFound
		}

		return reassigned, fmt.Errorf("%w: lens is used by pictures", ErrInUse)
	}

	lensCache.Delete(m.LensSlug)

	// Content channels carry only stable identities, never entity fields; publish the slug.
	event.EntitiesDeleted("lenses", []string{m.LensSlug})

	event.Publish("count.lenses", event.Data{
		"count": -1,
	})

	return reassigned, nil
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
	} else if m.Unknown() {
		// Renaming the shared placeholder would relabel every picture without lens information.
		return fmt.Errorf("unknown lens cannot be changed")
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
