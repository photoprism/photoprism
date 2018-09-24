package models

import (
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

type Camera struct {
	gorm.Model
	CameraSlug        string
	CameraModel       string
	CameraMake        string
	CameraType        string
	CameraOwner       string
	CameraDescription string `gorm:"type:text;"`
	CameraNotes       string `gorm:"type:text;"`
}

func NewCamera(modelName string, makeName string) *Camera {
	if modelName == "" {
		modelName = "Unknown"
	}

	cameraSlug := slug.MakeLang(modelName, "en")

	result := &Camera{
		CameraModel: modelName,
		CameraMake:  makeName,
		CameraSlug:  cameraSlug,
	}

	return result
}

func (c *Camera) FirstOrCreate(db *gorm.DB) *Camera {
	db.FirstOrCreate(c, "camera_model = ? AND camera_make = ?", c.CameraModel, c.CameraMake)

	return c
}
