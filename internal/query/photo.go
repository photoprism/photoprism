package query

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
)

// PhotoByID returns a Photo based on the ID.
func PhotoByID(photoID uint64) (photo entity.Photo, err error) {
	if err := UnscopedDb().Where("id = ?", photoID).
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		Preload("Camera").
		Preload("Lens").
		Preload("Details").
		Preload("Place").
		Preload("Cell").
		Preload("Cell.Place").
		First(&photo).Error; err != nil {
		return photo, err
	}

	return photo, nil
}

// PhotoByUID returns a Photo based on the UID.
func PhotoByUID(photoUID string) (photo entity.Photo, err error) {
	if err := UnscopedDb().Where("photo_uid = ?", photoUID).
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		Preload("Camera").
		Preload("Lens").
		Preload("Details").
		Preload("Place").
		Preload("Cell").
		Preload("Cell.Place").
		First(&photo).Error; err != nil {
		return photo, err
	}

	return photo, nil
}

// PhotoPreloadByUID returns a Photo based on the UID with all dependencies preloaded.
func PhotoPreloadByUID(photoUID string) (photo entity.Photo, err error) {
	if err := UnscopedDb().Where("photo_uid = ?", photoUID).
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		Preload("Camera").
		Preload("Lens").
		Preload("Details").
		Preload("Place").
		Preload("Cell").
		Preload("Cell.Place").
		First(&photo).Error; err != nil {
		return photo, err
	}

	photo.PreloadMany()

	return photo, nil
}

// PhotosMissing returns photo entities without existing files.
func PhotosMissing(limit int, offset int) (entities entity.Photos, err error) {
	err = Db().
		Select("photos.*").
		Where("id NOT IN (SELECT photo_id FROM files WHERE file_missing = 0 AND file_root = '/' AND deleted_at IS NULL)").
		Where("photos.photo_type <> ?", entity.TypeText).
		Group("photos.id").
		Limit(limit).Offset(offset).Find(&entities).Error

	return entities, err
}

// ResetPhotoQuality sets photo quality scores to -1 if files are missing.
func ResetPhotoQuality() error {
	log.Info("index: flagging hidden files")

	return Db().Table("photos").
		Where("id NOT IN (SELECT photo_id FROM files WHERE file_primary = 1 AND file_missing = 0 AND deleted_at IS NULL)").
		Update("photo_quality", -1).Error
}

// PhotosCheck returns photos selected for maintenance.
func PhotosCheck(limit, offset int, delay time.Duration) (entities entity.Photos, err error) {
	err = Db().
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		Preload("Camera").
		Preload("Lens").
		Preload("Details").
		Preload("Place").
		Preload("Cell").
		Preload("Cell.Place").
		Where("checked_at IS NULL OR checked_at < ?", time.Now().Add(-1*time.Hour*24*3)).
		Where("updated_at < ? OR (cell_id = 'zz' AND photo_lat <> 0)", time.Now().Add(-1*delay)).
		Order("photos.ID ASC").Limit(limit).Offset(offset).Find(&entities).Error

	return entities, err
}

// OrphanPhotos finds orphan index entries that may be removed.
func OrphanPhotos() (photos entity.Photos, err error) {
	err = UnscopedDb().
		Raw(`SELECT * FROM photos WHERE 
			deleted_at IS NOT NULL 
			AND photo_quality = -1 
			AND id NOT IN (SELECT photo_id FROM files WHERE files.deleted_at IS NULL)`).
		Find(&photos).Error

	return photos, err
}

// FixPrimaries tries to set a primary file for photos that have none.
func FixPrimaries() error {
	log.Info("index: updating primary files")

	var photos entity.Photos

	if err := UnscopedDb().
		Raw(`SELECT * FROM photos 
			WHERE deleted_at IS NULL 
			AND id NOT IN (SELECT photo_id FROM files WHERE file_primary = 1)`).
		Find(&photos).Error; err != nil {
		return err
	}

	for _, p := range photos {
		log.Debugf("index: finding new primary file for photo %s", p.PhotoUID)

		if err := p.SetPrimary(""); err != nil {
			log.Infof("index: %s (set primary)", err)
		}
	}

	return nil
}
