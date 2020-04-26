package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/mutex"
)

// Camera model and make (as extracted from UpdateExif metadata)
type Camera struct {
	ID                uint   `gorm:"primary_key"`
	CameraSlug        string `gorm:"type:varbinary(128);unique_index;"`
	CameraModel       string `gorm:"type:varchar(128);"`
	CameraMake        string `gorm:"type:varchar(128);"`
	CameraType        string `gorm:"type:varchar(128);"`
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
func CreateUnknownCamera(db *gorm.DB) {
	UnknownCamera.FirstOrCreate(db)
}

// NewCamera creates a camera entity from a model name and a make name.
func NewCamera(modelName string, makeName string) *Camera {
	makeName = strings.TrimSpace(makeName)

	if modelName == "" {
		return &UnknownCamera
	} else if strings.HasPrefix(modelName, makeName) {
		modelName = strings.TrimSpace(modelName[len(makeName):])
	}

	var cameraSlug string

	if makeName != "" {
		cameraSlug = slug.Make(makeName + " " + modelName)
	} else {
		cameraSlug = slug.Make(modelName)
	}

	result := &Camera{
		CameraModel: modelName,
		CameraMake:  makeName,
		CameraSlug:  cameraSlug,
	}

	return result
}

// FirstOrCreate checks wether the camera model exist already in the database
func (m *Camera) FirstOrCreate(db *gorm.DB) *Camera {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "camera_model = ? AND camera_make = ?", m.CameraModel, m.CameraMake).Error; err != nil {
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
