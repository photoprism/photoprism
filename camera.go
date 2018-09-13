package photoprism

import (
	"github.com/jinzhu/gorm"
)

type Camera struct {
	gorm.Model
	CameraModel string
	CameraType  string
	CameraNotes string
}

func NewCamera(modelName string) *Camera {
	if modelName == "" {
		modelName = "Unknown"
	}

	result := &Camera{
		CameraModel: modelName,
	}

	return result
}

func (c *Camera) FirstOrCreate(db *gorm.DB) *Camera {
	db.FirstOrCreate(c, "camera_model = ?", c.CameraModel)

	return c
}
