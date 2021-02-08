package entity

import (
	"strings"
	"sync"
	"time"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"
)

var cameraMutex = sync.Mutex{}

// Camera model and make (as extracted from UpdateExif metadata)
type Camera struct {
	ID                uint       `gorm:"primary_key" json:"ID" yaml:"ID"`
	CameraSlug        string     `gorm:"type:VARBINARY(255);unique_index;" json:"Slug" yaml:"-"`
	CameraName        string     `gorm:"type:VARCHAR(255);" json:"Name" yaml:"Name"`
	CameraMake        string     `gorm:"type:VARCHAR(255);" json:"Make" yaml:"Make,omitempty"`
	CameraModel       string     `gorm:"type:VARCHAR(255);" json:"Model" yaml:"Model,omitempty"`
	CameraType        string     `gorm:"type:VARCHAR(255);" json:"Type,omitempty" yaml:"Type,omitempty"`
	CameraDescription string     `gorm:"type:TEXT;" json:"Description,omitempty" yaml:"Description,omitempty"`
	CameraNotes       string     `gorm:"type:TEXT;" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	CreatedAt         time.Time  `json:"-" yaml:"-"`
	UpdatedAt         time.Time  `json:"-" yaml:"-"`
	DeletedAt         *time.Time `sql:"index" json:"-" yaml:"-"`
}

var UnknownCamera = Camera{
	CameraSlug:  "zz",
	CameraName:  "Unknown",
	CameraMake:  "",
	CameraModel: "Unknown",
}

// CreateUnknownCamera initializes the database with an unknown camera if not exists
func CreateUnknownCamera() {
	FirstOrCreateCamera(&UnknownCamera)
}

// NewCamera creates a camera entity from a model name and a make name.
func NewCamera(modelName string, makeName string) *Camera {
	modelName = txt.Clip(modelName, txt.ClipDefault)
	makeName = txt.Clip(makeName, txt.ClipDefault)

	if modelName == "" && makeName == "" {
		return &UnknownCamera
	} else if strings.HasPrefix(modelName, makeName) {
		modelName = strings.TrimSpace(modelName[len(makeName):])
	}

	if n, ok := CameraMakes[makeName]; ok {
		makeName = n
	}

	if n, ok := CameraModels[modelName]; ok {
		modelName = n
	}

	var name []string

	if makeName != "" {
		name = append(name, makeName)
	}

	if modelName != "" {
		name = append(name, modelName)
	}

	cameraName := strings.Join(name, " ")
	cameraSlug := slug.Make(txt.Clip(cameraName, txt.ClipSlug))

	result := &Camera{
		CameraSlug:  cameraSlug,
		CameraName:  cameraName,
		CameraMake:  makeName,
		CameraModel: modelName,
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
		log.Debugf("camera: cache hit for %s", m.CameraSlug)

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
	} else if res := Db().Where("camera_slug = ?", m.CameraSlug).First(&result); res.Error == nil {
		cameraCache.SetDefault(m.CameraSlug, &result)
		return &result
	} else {
		log.Errorf("camera: %s (create %s)", err.Error(), txt.Quote(m.String()))
	}

	return &UnknownCamera
}

// String returns an identifier that can be used in logs.
func (m *Camera) String() string {
	return m.CameraName
}

// Unknown returns true if the camera is not a known make or model.
func (m *Camera) Unknown() bool {
	return m.CameraSlug == "" || m.CameraSlug == UnknownCamera.CameraSlug
}
