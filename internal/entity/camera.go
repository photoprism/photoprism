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

var cameraMutex = sync.Mutex{}

// Cameras represents a list of cameras.
type Cameras []Camera

// Camera model and make (as extracted from UpdateExif metadata)
type Camera struct {
	ID                uint       `gorm:"primary_key" json:"ID" yaml:"ID"`
	CameraSlug        string     `gorm:"type:VARBINARY(160);unique_index;" json:"Slug" yaml:"-"`
	CameraName        string     `gorm:"type:VARCHAR(160);" json:"Name" yaml:"Name"`
	CameraMake        string     `gorm:"type:VARCHAR(160);" json:"Make" yaml:"Make,omitempty"`
	CameraModel       string     `gorm:"type:VARCHAR(160);" json:"Model" yaml:"Model,omitempty"`
	CameraType        string     `gorm:"type:VARCHAR(100);" json:"Type,omitempty" yaml:"Type,omitempty"`
	CameraDescription string     `gorm:"type:VARCHAR(2048);" json:"Description,omitempty" yaml:"Description,omitempty"`
	CameraNotes       string     `gorm:"type:VARCHAR(1024);" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	CameraSrc         string     `gorm:"type:VARBINARY(8);default:'';" json:"-" yaml:"-"`
	CreatedAt         time.Time  `json:"-" yaml:"-"`
	UpdatedAt         time.Time  `json:"-" yaml:"-"`
	DeletedAt         *time.Time `sql:"index" json:"-" yaml:"-"`
}

// TableName returns the entity table name.
func (Camera) TableName() string {
	return "cameras"
}

// UnknownCamera is the placeholder used when no camera make or model is known.
var UnknownCamera = Camera{
	CameraSlug:  UnknownID,
	CameraName:  "Unknown",
	CameraMake:  MakeNone,
	CameraModel: ModelUnknown,
}

// CreateUnknownCamera initializes the database with an unknown camera if not exists
func CreateUnknownCamera() {
	UnknownCamera = *FirstOrCreateCamera(&UnknownCamera)
}

// NewCamera creates a new camera entity from make and model names.
func NewCamera(makeName string, modelName string) *Camera {
	makeName = strings.TrimSpace(makeName)
	modelName = strings.Trim(modelName, " \t\r\n-_")

	if modelName == "" && makeName == "" {
		return &UnknownCamera
	}

	modelName = trimMakePrefix(modelName, makeName)

	// Normalize make name.
	if n, ok := CameraMakes[makeName]; ok {
		makeName = n
	}

	// Normalize model name.
	if n, ok := CameraModels[modelName]; ok {
		modelName = n
	}

	modelName = trimMakePrefix(modelName, makeName)

	// Determine device type based on make and model.
	cameraType := GetCameraType(makeName, modelName)

	var name []string

	if makeName != "" {
		name = append(name, makeName)
	}

	if modelName != "" {
		name = append(name, modelName)
	}

	cameraName := strings.Join(name, " ")

	result := &Camera{
		CameraSlug:  txt.Slug(cameraName),
		CameraName:  txt.Clip(cameraName, txt.ClipName),
		CameraMake:  txt.Clip(makeName, txt.ClipName),
		CameraModel: txt.Clip(modelName, txt.ClipName),
		CameraType:  cameraType,
	}

	return result
}

// Create inserts a new row to the database.
func (m *Camera) Create() error {
	cameraMutex.Lock()
	defer cameraMutex.Unlock()

	return Db().Create(m).Error
}

// FirstOrCreateCamera returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateCamera(m *Camera) *Camera {
	if m.CameraSlug == "" {
		return &UnknownCamera
	}

	if cacheData, ok := cameraCache.Get(m.CameraSlug); ok {
		log.Tracef("camera: cache hit for %s", m.CameraSlug)

		return cacheData.(*Camera)
	}

	result := Camera{}

	if res := Db().Where("camera_slug = ?", m.CameraSlug).First(&result); res.Error == nil {
		cameraCache.SetDefault(m.CameraSlug, &result)
		return &result
	} else if err := m.Create(); err == nil {
		if !m.Unknown() {
			// Content channels carry only stable identities, never entity fields; publish the slug.
			event.EntitiesCreated("cameras", []string{m.CameraSlug})

			event.Publish("count.cameras", event.Data{
				"count": 1,
			})
		}

		cameraCache.SetDefault(m.CameraSlug, m)

		return m
	} else if res = Db().Where("camera_slug = ?", m.CameraSlug).First(&result); res.Error == nil {
		cameraCache.SetDefault(m.CameraSlug, &result)
		return &result
	} else {
		log.Errorf("camera: %s (create %s)", err.Error(), clean.Log(m.String()))
	}

	return &UnknownCamera
}

// AddCamera returns the camera with the specified make and model, and creates it if it does not exist yet.
// The camera is marked as added manually, so that purging orphans keeps it while no picture references it.
func AddCamera(makeName, modelName string) (result *Camera, created bool, err error) {
	makeName = strings.TrimSpace(makeName)
	modelName = strings.TrimSpace(modelName)

	if makeName == "" || modelName == "" {
		return nil, false, fmt.Errorf("%w: make and model must not be empty", ErrInvalidValue)
	}

	m := NewCamera(makeName, modelName)

	if strings.TrimSpace(m.CameraModel) == "" {
		return nil, false, fmt.Errorf("%w: model must not be empty after removing the make", ErrInvalidValue)
	} else if m.Unknown() {
		return nil, false, fmt.Errorf("%w: make and model must not match the unknown camera", ErrInvalidValue)
	}

	// Report an existing camera instead of creating a duplicate, unless it has been purged in the meantime.
	if existing := lookupExistingCamera(m); existing != nil && !existing.Unknown() {
		if err = existing.markManual(); err == nil {
			return existing, false, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, err
		}
	}

	m.CameraSrc = SrcManual

	if result = FirstOrCreateCamera(m); result == m {
		return result, true, nil
	} else if result.Unknown() {
		return nil, false, fmt.Errorf("failed to create camera %s", m.String())
	} else if err = result.markManual(); err != nil {
		return nil, false, err
	}

	return result, false, nil
}

// lookupExistingCamera finds an existing camera when adding one, and can be replaced in tests to simulate a concurrent purge.
var lookupExistingCamera = findExistingCamera

// findExistingCamera returns the camera with the same slug, or else with the same make and model, if any.
// The make and model lookup covers renamed records, whose slug no longer matches their name.
func findExistingCamera(m *Camera) *Camera {
	existing := Camera{}

	if Db().Where("camera_slug = ?", m.CameraSlug).First(&existing).Error == nil {
		return &existing
	}

	if found := findCamerasByMakeModel(m.CameraMake, m.CameraModel); len(found) > 0 {
		return &found[0]
	}

	return nil
}

// FindCamerasByMakeModel returns the cameras that match the make and model as stored, or else after normalizing them.
// It ignores the slug, which a renamed record keeps and may therefore belong to another name.
func FindCamerasByMakeModel(makeName, modelName string) Cameras {
	makeName = strings.TrimSpace(makeName)
	modelName = strings.TrimSpace(modelName)

	if makeName == "" || modelName == "" {
		return nil
	} else if found := findCamerasByMakeModel(makeName, modelName); len(found) > 0 {
		// Records saved with a different normalization, e.g. by an older version, match as stored.
		return found
	}

	m := NewCamera(makeName, modelName)

	if m.Unknown() || strings.TrimSpace(m.CameraModel) == "" {
		return nil
	}

	return findCamerasByMakeModel(m.CameraMake, m.CameraModel)
}

// findCamerasByMakeModel returns the cameras with exactly the specified make and model, ordered by ID.
func findCamerasByMakeModel(makeName, modelName string) (result Cameras) {
	var candidates Cameras

	if Db().Where("camera_make = ? AND camera_model = ?", makeName, modelName).Order("id").Find(&candidates).Error != nil {
		return nil
	}

	// Compare in Go, since MariaDB collations also match values that differ in case, accents, or emoji.
	for i := range candidates {
		if candidates[i].CameraMake == makeName && candidates[i].CameraModel == modelName && !candidates[i].Unknown() {
			result = append(result, candidates[i])
		}
	}

	return result
}

// markManual marks an existing camera as added manually, so that purging orphans keeps it.
// It returns gorm.ErrRecordNotFound if the camera no longer exists, e.g. because it has been purged.
func (m *Camera) markManual() error {
	if m.ID == 0 {
		// Without a primary key, the update would apply to all rows.
		return fmt.Errorf("empty id")
	}

	cameraMutex.Lock()
	defer cameraMutex.Unlock()

	if res := UnscopedDb().Model(m).UpdateColumn("camera_src", SrcManual); res.Error != nil {
		return res.Error
	} else if res.RowsAffected == 0 {
		// MariaDB counts only changed rows, so check whether the camera still exists.
		var count int

		if err := UnscopedDb().Model(&Camera{}).Where("id = ?", m.ID).Count(&count).Error; err != nil {
			return err
		} else if count == 0 {
			return gorm.ErrRecordNotFound
		}
	}

	m.CameraSrc = SrcManual
	cameraCache.SetDefault(m.CameraSlug, m)

	return nil
}

// unknownCameraID returns the ID of the unknown camera placeholder, and initializes the placeholder if needed.
func unknownCameraID() (uint, error) {
	if UnknownCamera.ID == 0 {
		CreateUnknownCamera()
	}

	if UnknownCamera.ID == 0 {
		return 0, fmt.Errorf("unknown camera not found")
	}

	return UnknownCamera.ID, nil
}

// PhotoCount returns the number of pictures that reference the camera, including archived and deleted pictures.
func (m *Camera) PhotoCount() (count int, err error) {
	if m.ID == 0 {
		return 0, fmt.Errorf("empty id")
	}

	err = UnscopedDb().Model(&Photo{}).Where("camera_id = ?", m.ID).Count(&count).Error

	return count, err
}

// Delete permanently removes the camera, and first assigns the pictures that reference it to the unknown camera if requested.
// It returns ErrInUse if pictures still reference the camera, so that none of them is left with a dangling reference.
func (m *Camera) Delete(reassign bool) (reassigned int64, err error) {
	if m.ID == 0 {
		return 0, fmt.Errorf("empty id")
	}

	// Resolve the placeholder from the database, since the CLI does not initialize it on startup.
	unknownID, err := unknownCameraID()

	if err != nil {
		return 0, err
	} else if m.Unknown() || m.ID == unknownID {
		return 0, fmt.Errorf("%w: unknown camera cannot be deleted", ErrInvalidValue)
	}

	cameraMutex.Lock()
	defer cameraMutex.Unlock()

	if reassign {
		res := UnscopedDb().Model(&Photo{}).Where("camera_id = ?", m.ID).UpdateColumn("camera_id", unknownID)

		if res.Error != nil {
			return 0, res.Error
		}

		reassigned = res.RowsAffected
	}

	// Delete the camera only if no picture references it, e.g. after being assigned to it by a concurrent index run.
	res := UnscopedDb().Exec(`DELETE FROM cameras WHERE id = ? AND NOT EXISTS (SELECT 1 FROM photos WHERE photos.camera_id = cameras.id)`, m.ID)

	if res.Error != nil {
		return reassigned, res.Error
	} else if res.RowsAffected == 0 {
		var count int

		if err = UnscopedDb().Model(&Camera{}).Where("id = ?", m.ID).Count(&count).Error; err != nil {
			return reassigned, err
		} else if count == 0 {
			return reassigned, gorm.ErrRecordNotFound
		}

		return reassigned, fmt.Errorf("%w: camera is used by pictures", ErrInUse)
	}

	cameraCache.Delete(m.CameraSlug)

	// Content channels carry only stable identities, never entity fields; publish the slug.
	event.EntitiesDeleted("cameras", []string{m.CameraSlug})

	event.Publish("count.cameras", event.Data{
		"count": -1,
	})

	return reassigned, nil
}

// String returns an identifier that can be used in logs.
func (m *Camera) String() string {
	if m == nil {
		return "Camera<nil>"
	}

	return clean.Log(m.CameraName)
}

// Scanner checks whether the model appears to be a scanner.
func (m *Camera) Scanner() bool {
	switch m.CameraType {
	case CameraTypeFilm, CameraTypeScanner:
		return true
	}

	if m.CameraSlug == "" {
		return false
	}

	return strings.Contains(m.CameraSlug, "scan")
}

// Mobile checks whether the model appears to be a mobile device.
func (m *Camera) Mobile() bool {
	switch m.CameraType {
	case CameraTypePhone, CameraTypeTablet:
		return true
	default:
		return false
	}
}

// Unknown returns true if the camera is not a known make or model.
func (m *Camera) Unknown() bool {
	return m.CameraSlug == "" || m.CameraSlug == UnknownCamera.CameraSlug
}

// UpdateMakeModel updates the make and model of an existing camera, e.g. to fix entries that
// ExifTool decodes with a missing or garbled make.
// The camera slug is intentionally left unchanged so existing photo references and the unique slug
// index are preserved across renames.
func (m *Camera) UpdateMakeModel(makeName, modelName string) error {
	if m.ID == 0 {
		return fmt.Errorf("empty id")
	} else if m.Unknown() {
		// Renaming the shared placeholder would relabel every picture without camera information.
		return fmt.Errorf("unknown camera cannot be changed")
	}

	makeName = strings.TrimSpace(makeName)
	modelName = strings.TrimSpace(modelName)

	if makeName == "" || modelName == "" {
		return fmt.Errorf("make and model must not be empty")
	}

	cam := NewCamera(makeName, modelName)
	// Override the changeable fields.
	m.CameraMake = cam.CameraMake
	m.CameraModel = cam.CameraModel
	m.CameraName = cam.CameraName
	m.CameraType = cam.CameraType

	cameraMutex.Lock()
	defer cameraMutex.Unlock()
	if err := Db().Save(m).Error; err != nil {
		return err
	} else {
		if !m.Unknown() {
			event.EntitiesUpdated("cameras", []string{m.CameraSlug})
		}
		cameraCache.SetDefault(m.CameraSlug, m)
	}
	return nil
}

// SaveForm validates the form, copies its data into the camera, and persists it.
func (m *Camera) SaveForm(f *form.Camera) error {
	if f == nil {
		return fmt.Errorf("form is nil")
	} else if err := f.Validate(); err != nil {
		return err
	}

	if err := deepcopier.Copy(m).From(f); err != nil {
		return err
	}

	return m.UpdateMakeModel(f.CameraMake, f.CameraModel)
}
