package models

import (
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

// Camera model and make (as extracted from EXIF metadata)
type Camera struct {
	Model
	CameraSlug        string
	CameraModel       string
	CameraMake        string
	CameraType        string
	CameraOwner       string
	CameraDescription string `gorm:"type:text;"`
	CameraNotes       string `gorm:"type:text;"`
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
	db.FirstOrCreate(m, "camera_model = ? AND camera_make = ?", m.CameraModel, m.CameraMake)

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
