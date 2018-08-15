package photoprism

import (
	"github.com/jinzhu/gorm"
)

type Camera struct {
	gorm.Model
	ModelName string
	Type      string
	Notes     string
}

func NewCamera(modelName string) *Camera {
	if modelName == "" {
		modelName = "Unknown"
	}

	result := &Camera{
		ModelName: modelName,
	}

	return result
}

func (c *Camera) FirstOrCreate(db *gorm.DB) *Camera {
	db.FirstOrCreate(c, "model_name = ?", c.ModelName)

	return c
}
