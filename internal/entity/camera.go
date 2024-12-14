package entity

import (
	"strings"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/ulule/deepcopier"
	"gorm.io/gorm"
)

var cameraMutex = sync.Mutex{}

// Cameras represents a list of cameras.
type Cameras []Camera

// Camera model and make (as extracted from UpdateExif metadata)
type Camera struct {
	ID                uint           `gorm:"primaryKey;" json:"ID" yaml:"ID"`
	CameraSlug        string         `gorm:"type:bytes;size:160;uniqueIndex;" json:"Slug" yaml:"-"`
	CameraName        string         `gorm:"size:160;" json:"Name" yaml:"Name"`
	CameraMake        string         `gorm:"size:160;" json:"Make" yaml:"Make,omitempty"`
	CameraModel       string         `gorm:"size:160;" json:"Model" yaml:"Model,omitempty"`
	CameraType        string         `gorm:"size:100;" json:"Type,omitempty" yaml:"Type,omitempty"`
	CameraDescription string         `gorm:"size:2048;" json:"Description,omitempty" yaml:"Description,omitempty"`
	CameraNotes       string         `gorm:"size:1024;" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	CreatedAt         time.Time      `json:"-" yaml:"-"`
	UpdatedAt         time.Time      `json:"-" yaml:"-"`
	DeletedAt         gorm.DeletedAt `sql:"index" json:"-" yaml:"-"`
}

// TableName returns the entity table name.
func (Camera) TableName() string {
	return "cameras"
}

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
			event.EntitiesCreated("cameras", []*Camera{m})

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

// ScopedSearchFirstCamera populates camera with the results of a Where(query, values) excluding soft delete records
func ScopedSearchFirstCamera(camera *Camera, query string, values ...interface{}) (tx *gorm.DB) {
	// Preload related entities if a matching record is found.
	stmt := Db()
	tempCamera := &Camera{}
	if tx = stmt.Where(query, values...).First(tempCamera); tx.Error == nil {
		deepcopier.Copy(tempCamera).To(camera)
	}
	return tx
}
