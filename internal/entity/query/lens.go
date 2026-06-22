package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// FindLensBySlug returns an existing entity if exists.
func FindLensBySlug(slug string) *entity.Lens {
	if slug == "" {
		return nil
	}

	l := entity.Lens{}

	if err := Db().Where("lens_slug = ?", slug).First(&l).Error; err != nil {
		return nil
	}

	return &l
}

// FindLensByID returns an existing entity if exists.
func FindLensByID(id uint) *entity.Lens {
	if id == 0 {
		return nil
	}

	l := entity.Lens{}

	if err := Db().Where("id = ?", id).First(&l).Error; err != nil {
		return nil
	}

	return &l
}
