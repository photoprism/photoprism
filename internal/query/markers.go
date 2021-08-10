package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// MarkerByID returns a Marker based on the ID.
func MarkerByID(id uint) (marker entity.Marker, err error) {
	if err := UnscopedDb().Where("id = ?", id).
		First(&marker).Error; err != nil {
		return marker, err
	}

	return marker, nil
}
