package entity

import "gorm.io/gorm"

// Photos represents a list of photos.
type Photos []Photo

// Photos returns the result as a slice of Photo.
func (m Photos) Photos() []PhotoInterface {
	result := make([]PhotoInterface, len(m))

	for i := range m {
		result[i] = &m[i]
	}

	return result
}

// UIDs returns tbe photo UIDs as string slice.
func (m Photos) UIDs() []string {
	result := make([]string, len(m))

	for i, photo := range m {
		result[i] = photo.GetUID()
	}

	return result
}

// UnscopedSearchPhotos populates the photos that match the results of a Where(query, values) including soft delete records
func UnscopedSearchPhotos(photos *Photos, query string, values ...interface{}) (tx *gorm.DB) {
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
func ScopedSearchPhotos(photos *Photos, query string, values ...interface{}) (tx *gorm.DB) {
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
