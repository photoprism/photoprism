package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Camera model and make (as extracted from UpdateExif metadata)
type Camera struct {
	ID                uint   `gorm:"primary_key"`
	CameraSlug        string `gorm:"type:varbinary(255);unique_index;"`
	CameraModel       string `gorm:"type:varchar(255);"`
	CameraMake        string `gorm:"type:varchar(255);"`
	CameraType        string `gorm:"type:varchar(255);"`
	CameraDescription string `gorm:"type:text;"`
	CameraNotes       string `gorm:"type:text;"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time `sql:"index"`
}

var UnknownCamera = Camera{
	CameraModel: "Unknown",
	CameraMake:  "",
	CameraSlug:  "zz",
}

// CreateUnknownCamera initializes the database with an unknown camera if not exists
func CreateUnknownCamera() {
	UnknownCamera.FirstOrCreate()
}

// NewCamera creates a camera entity from a model name and a make name.
func NewCamera(modelName string, makeName string) *Camera {
	modelName = txt.Clip(modelName, txt.ClipDefault)
	makeName = txt.Clip(makeName, txt.ClipDefault)

	if modelName == "" {
		return &UnknownCamera
	} else if strings.HasPrefix(modelName, makeName) {
		modelName = strings.TrimSpace(modelName[len(makeName):])
	}

	var cameraSlug string

	if makeName != "" {
		cameraSlug = slug.Make(txt.Clip(makeName+" "+modelName, txt.ClipSlug))
	} else {
		cameraSlug = slug.Make(txt.Clip(modelName, txt.ClipSlug))
	}

	result := &Camera{
		CameraModel: modelName,
		CameraMake:  makeName,
		CameraSlug:  cameraSlug,
	}

	return result
}

// FirstOrCreate checks if the camera model exist already in the database
func (m *Camera) FirstOrCreate() *Camera {
	if err := Db().FirstOrCreate(m, "camera_model = ? AND camera_make = ?", m.CameraModel, m.CameraMake).Error; err != nil {
		log.Errorf("camera: %s", err)
	}

	return m
}

// String returns a string designing the given Camera entity
func (m *Camera) String() string {
	if m.CameraMake != "" && m.CameraModel != "" {
		return fmt.Sprintf("%s %s", m.CameraMake, m.CameraModel)
	} else if m.CameraModel != "" {
		return m.CameraModel
	}

	return ""
}
