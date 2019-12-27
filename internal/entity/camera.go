package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

// Camera model and make (as extracted from UpdateExif metadata)
type Camera struct {
	ID                uint `gorm:"primary_key"`
	CameraSlug        string
	CameraModel       string
	CameraMake        string
	CameraType        string
	CameraOwner       string
	CameraDescription string `gorm:"type:text;"`
	CameraNotes       string `gorm:"type:text;"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time `sql:"index"`
}

func NewCamera(modelName string, makeName string) *Camera {
	makeName = strings.TrimSpace(makeName)

	if modelName == "" {
		modelName = "Unknown"
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

func (m *Camera) FirstOrCreate(db *gorm.DB) *Camera {
	if err := db.FirstOrCreate(m, "camera_model = ? AND camera_make = ?", m.CameraModel, m.CameraMake).Error; err != nil {
		log.Errorf("camera: %s", err)
	}

	return m
}

func (m *Camera) String() string {
	if m.CameraMake != "" && m.CameraModel != "" {
		return fmt.Sprintf("%s %s", m.CameraMake, m.CameraModel)
	} else if m.CameraModel != "" {
		return m.CameraModel
	}

	return ""
}
