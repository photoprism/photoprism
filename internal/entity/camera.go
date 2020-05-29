package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Camera model and make (as extracted from UpdateExif metadata)
type Camera struct {
	ID                uint       `gorm:"primary_key" json:"ID" yaml:"ID"`
	CameraSlug        string     `gorm:"type:varbinary(255);unique_index;" json:"Slug" yaml:"-"`
	CameraName        string     `gorm:"type:varchar(255);" json:"Name" yaml:"Name"`
	CameraMake        string     `gorm:"type:varchar(255);" json:"Make" yaml:"Make,omitempty"`
	CameraModel       string     `gorm:"type:varchar(255);" json:"Model" yaml:"Model,omitempty"`
	CameraType        string     `gorm:"type:varchar(255);" json:"Type,omitempty" yaml:"Type,omitempty"`
	CameraDescription string     `gorm:"type:text;" json:"Description,omitempty" yaml:"Description,omitempty"`
	CameraNotes       string     `gorm:"type:text;" json:"Notes,omitempty" yaml:"Notes,omitempty"`
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

var CameraMakes = map[string]string{
	"OLYMPUS OPTICAL CO.,LTD": "Olympus",
}

var CameraModels = map[string]string{
	"ELE-L29":  "P30",
	"ELE-AL00": "P30",
	"ELE-L04":  "P30",
	"ELE-L09":  "P30",
	"ELE-TL00": "P30",
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
	return Db().Create(m).Error
}

// FirstOrCreateCamera returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateCamera(m *Camera) *Camera {
	result := Camera{}

	if err := Db().Where("camera_model = ? AND camera_make = ?", m.CameraModel, m.CameraMake).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Errorf("camera: %s", err)
		return nil
	}

	if !m.Unknown() {
		event.EntitiesCreated("cameras", []*Camera{m})

		event.Publish("count.cameras", event.Data{
			"count": 1,
		})
	}

	return m
}

// String returns a string designing the given Camera entity
func (m *Camera) String() string {
	return m.CameraName
}

// Unknown returns true if the camera is not a known make or model.
func (m *Camera) Unknown() bool {
	return m.CameraSlug == "" || m.CameraSlug == UnknownCamera.CameraSlug
}
