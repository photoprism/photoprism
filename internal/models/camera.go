package models

import (
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

type Camera struct {
	gorm.Model
	CameraSlug  string
	CameraModel string
	CameraType  string
	CameraNotes string
}

func NewCamera(modelName string) *Camera {
	if modelName == "" {
		modelName = "Unknown"
	}

	cameraSlug := slug.MakeLang(modelName, "en")

	result := &Camera{
		CameraModel: modelName,
		CameraSlug: cameraSlug,
	}

	return result
}

func (c *Camera) FirstOrCreate(db *gorm.DB) *Camera {
	db.FirstOrCreate(c, "camera_model = ?", c.CameraModel)

	return c
}
