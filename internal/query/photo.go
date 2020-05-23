package query

import (
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
)

// PhotoByID returns a Photo based on the ID.
func PhotoByID(photoID uint64) (photo entity.Photo, err error) {
	if err := UnscopedDb().Where("id = ?", photoID).
		Preload("Links").
		Preload("Camera").
		Preload("Lens").
		Preload("Details").
		Preload("Location").
		Preload("Location.Place").
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		First(&photo).Error; err != nil {
		return photo, err
	}

	return photo, nil
}

// PhotoByUID returns a Photo based on the UID.
func PhotoByUID(photoUID string) (photo entity.Photo, err error) {
	if err := UnscopedDb().Where("photo_uid = ?", photoUID).
		Preload("Links").
		Preload("Camera").
		Preload("Lens").
		Preload("Details").
		Preload("Location").
		Preload("Location.Place").
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		First(&photo).Error; err != nil {
		return photo, err
	}

	return photo, nil
}

// PreloadPhotoByUID returns a Photo based on the UID with all dependencies preloaded.
func PreloadPhotoByUID(photoUID string) (photo entity.Photo, err error) {
	if err := UnscopedDb().Where("photo_uid = ?", photoUID).
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		Preload("Camera").
		Preload("Lens").
		Preload("Links").
		Preload("Location").
		Preload("Location.Place").
		Preload("Details").
		First(&photo).Error; err != nil {
		return photo, err
	}

	photo.PreloadMany()

	return photo, nil
}

// MissingPhotos returns photo entities without existing files.
func MissingPhotos(limit int, offset int) (entities []entity.Photo, err error) {
	err = Db().
		Select("photos.*").
		Joins("JOIN files a ON photos.id = a.photo_id ").
		Joins("LEFT JOIN files b ON a.photo_id = b.photo_id AND a.id != b.id AND b.file_missing = 0").
		Where("a.file_missing = 1 AND b.id IS NULL").
		Where("photos.photo_type <> ?", entity.TypeText).
		Group("photos.id").
		Limit(limit).Offset(offset).Find(&entities).Error

	return entities, err
}

// ResetPhotosQuality resets the quality of photos without primary file to -1.
func ResetPhotosQuality() error {
	return Db().Table("photos").
		Where("id IN (SELECT photos.id FROM photos LEFT JOIN files ON photos.id = files.photo_id AND files.file_primary = 1 WHERE files.id IS NULL GROUP BY photos.id)").
		Update("photo_quality", -1).Error
}
