package query

import (
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
)

// CountPhotos returns the total number of Photo records in the database, delete or otherwise.
func CountPhotos() (numberOfPhotos int, err error) {
	count := int64(0)
	err = UnscopedDb().Model(&entity.Photo{}).Count(&count).Error
	return int(count), err
}

// UnscopedSearchPhotos populates the photos that match the results of a Where(query, values) including soft delete records
func UnscopedSearchPhotos(photos *entity.Photos, query string, values ...any) (tx *gorm.DB) {
	// Preload related entities if a matching record is found.
	stmt := UnscopedDb().
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		Preload("Camera").
		Preload("Lens").
		Preload("Details").
		Preload("Place").
		Preload("Cell").
		Preload("Cell.Place")

	return stmt.Where(query, values...).Find(photos)
}

// ScopedSearchPhotos populates the photos that match the results of a Where(query, values) excluding soft delete records
func ScopedSearchPhotos(photos *entity.Photos, query string, values ...any) (tx *gorm.DB) {
	// Preload related entities if a matching record is found.
	stmt := Db().
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		Preload("Camera").
		Preload("Lens").
		Preload("Details").
		Preload("Place").
		Preload("Cell").
		Preload("Cell.Place")

	return stmt.Where(query, values...).Find(photos)
}
