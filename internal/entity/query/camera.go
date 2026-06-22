package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// FindCameraBySlug returns an existing entity if exists.
func FindCameraBySlug(slug string) *entity.Camera {
	if slug == "" {
		return nil
	}

	c := entity.Camera{}

	if err := Db().Where("camera_slug = ?", slug).First(&c).Error; err != nil {
		return nil
	}

	return &c
}

// FindCameraByID returns an existing entity if exists.
func FindCameraByID(id uint) *entity.Camera {
	if id == 0 {
		return nil
	}

	c := entity.Camera{}

	if err := Db().Where("id = ?", id).First(&c).Error; err != nil {
		return nil
	}

	return &c
}
