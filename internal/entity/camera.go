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
	} else if strings.HasPrefix(modelName, makeName) {
		modelName = strings.TrimSpace(modelName[len(makeName):])
	}

	// Normalize make name.
	if n, ok := CameraMakes[makeName]; ok {
		makeName = n
	}

	// Normalize model name.
	if n, ok := CameraModels[modelName]; ok {
		modelName = n
	}

	if strings.HasPrefix(modelName, makeName) {
		modelName = strings.TrimSpace(modelName[len(makeName):])
	}

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
