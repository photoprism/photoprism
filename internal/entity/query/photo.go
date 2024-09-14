package query

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
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

// MissingPhotos returns photo entities without existing files.
func MissingPhotos(limit int, offset int) (entities entity.Photos, err error) {
	err = Db().
		Select("photos.*").
		Where("id NOT IN (SELECT photo_id FROM files WHERE file_missing = 0 AND file_root = '/' AND deleted_at IS NULL)").
		Order("photos.id").
		Limit(limit).Offset(offset).Find(&entities).Error

	return entities, err
}

// ArchivedPhotos finds and returns archived photos.
func ArchivedPhotos(limit int, offset int) (entities entity.Photos, err error) {
	err = UnscopedDb().
		Select("photos.*").
		Where("photos.photo_quality > -1").
		Where("photos.deleted_at IS NOT NULL").
		Order("photos.id").
		Limit(limit).Offset(offset).Find(&entities).Error

	return entities, err
}

// PhotosMetadataUpdate returns photos selected for metadata maintenance.
func PhotosMetadataUpdate(limit, offset int, delay, interval time.Duration) (photos entity.Photos, err error) {
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
		Where("checked_at IS NULL OR checked_at < ?", time.Now().Add(-1*interval)).
		Where("updated_at < ?", time.Now().Add(-1*delay)).
		Order("photos.ID ASC").Limit(limit).Offset(offset).Find(&photos).Error

	return photos, err
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
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var photos entity.Photos

	// Remove primary file flag from broken or missing files.
	if err := UnscopedDb().Table(entity.File{}.TableName()).
		Where("(file_error <> '' OR file_missing = 1) AND file_primary <> 0").
		UpdateColumn("file_primary", 0).Error; err != nil {
		return err
	}

	// Find photos without primary file.
	if err := UnscopedDb().
		Raw(`SELECT * FROM photos 
			WHERE deleted_at IS NULL 
			AND id NOT IN (SELECT photo_id FROM files WHERE file_primary = 1)`).
		Find(&photos).Error; err != nil {
		return err
	}

	if len(photos) == 0 {
		log.Debugf("index: found no photos without primary file [%s]", time.Since(start))
		return nil
	}

	// Try to find matching primary files.
	for _, p := range photos {
		log.Debugf("index: searching primary file for %s", p.PhotoUID)

		if err := p.SetPrimary(""); err != nil {
			log.Infof("index: %s", err)
		}
	}

	log.Debugf("index: updated primary files [%s]", time.Since(start))

	return nil
}

// FlagHiddenPhotos sets the quality score of photos without valid primary file to -1.
func FlagHiddenPhotos() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	// Start time for logs.
	start := time.Now()

	// IDs of hidden photos.
	var hidden []uint

	// Number of updated records.
	affected := 0

	// Find and flag hidden photos.
	if err = Db().Table(entity.Photo{}.TableName()).
		Where("id NOT IN (SELECT photo_id FROM files WHERE file_primary = 1 AND file_missing = 0 AND file_error = '' AND deleted_at IS NULL) AND photo_quality > -1").
		Pluck("id", &hidden).Error; err != nil {
		// Find query failed.
		return err
	} else if found := len(hidden); found == 0 {
		// Nothing to update.
		return nil
	} else {
		// Update photos in batches to be compatible with SQLite.
		batchSize := BatchSize()

		for i := 0; i < len(hidden); i += batchSize {
			j := i + batchSize

			if j > len(hidden) {
				j = len(hidden)
			}

			// Next batch.
			ids := hidden[i:j]

			// Set photos.photo_quality = -1.
			if result := UnscopedDb().Table(entity.Photo{}.TableName()).
				Where("id IN (?) AND photo_quality > -1", ids).
				UpdateColumn("photo_quality", -1); result.Error != nil {
				// Failed to flag all hidden photos.
				log.Warnf("index: failed to flag %d photos as hidden", len(hidden)-affected)
				return fmt.Errorf("%s while flagging hidden photos", result.Error)
			} else if result.RowsAffected > 0 {
				affected += int(result.RowsAffected)
			} else {
				affected += len(ids)
			}
		}
	}

	// Log number of affected rows, if any.
	if affected > 0 {
		log.Infof("index: flagged %s as hidden [%s]", english.Plural(affected, "photo", "photos"), time.Since(start))
	}

	return nil
}
