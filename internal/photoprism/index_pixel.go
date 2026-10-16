package photoprism

import (
	"errors"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
)

var googlePixelReconcileMutex sync.Mutex

// reconcileGooglePixelPhotos combines previously separate Google Pixel Camera capture records during a forced reindex.
func reconcileGooglePixelPhotos(related RelatedFiles) error {
	if related.Main == nil {
		return nil
	}

	capture := FindGooglePixelCapture(related.Main)
	if capture == nil || !capture.ValidPair() {
		return nil
	}

	files := capture.Files()
	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		fileNames = append(fileNames, file.RootRelName())
	}

	googlePixelReconcileMutex.Lock()
	defer googlePixelReconcileMutex.Unlock()

	var indexedFiles []entity.File
	if err := entity.UnscopedDb().Where("file_root = ? AND file_name IN (?)", entity.RootOriginals, fileNames).Find(&indexedFiles).Error; err != nil {
		return err
	}

	photoIDs := make([]uint, 0, len(indexedFiles))
	seen := make(map[uint]bool, len(indexedFiles))
	for _, file := range indexedFiles {
		if file.PhotoID > 0 && !seen[file.PhotoID] {
			seen[file.PhotoID] = true
			photoIDs = append(photoIDs, file.PhotoID)
		}
	}

	if len(photoIDs) < 2 {
		return nil
	}

	var photos entity.Photos
	if err := entity.UnscopedDb().Where("id IN (?)", photoIDs).Order("id ASC").Find(&photos).Error; err != nil {
		return err
	} else if len(photos) < 2 {
		return nil
	}

	// Determine the canonical photo: prefer the photo associated with the primary file.
	var primaryPhotoID uint
	if primary := capture.Primary(); primary != nil {
		primaryRelName := primary.RootRelName()
		for _, f := range indexedFiles {
			if f.FileName == primaryRelName {
				primaryPhotoID = f.PhotoID
				break
			}
		}
	}

	var canonical *entity.Photo
	var duplicates entity.Photos
	if primaryPhotoID > 0 {
		duplicates = make(entity.Photos, 0, len(photos)-1)
		for _, p := range photos {
			if p.ID == primaryPhotoID {
				canonical = p
			} else {
				duplicates = append(duplicates, p)
			}
		}
	}
	if canonical == nil {
		canonical = photos[0]
		duplicates = photos[1:]
	}

	orderedPhotos := make(entity.Photos, 0, len(photos))
	orderedPhotos = append(orderedPhotos, canonical)
	orderedPhotos = append(orderedPhotos, duplicates...)

	state := googlePixelCaptureState(orderedPhotos)

	favorite, private, panorama := false, false, false
	title, titleSrc := canonical.PhotoTitle, canonical.TitleSrc
	caption, captionSrc := canonical.PhotoCaption, canonical.CaptionSrc

	for _, photo := range photos {
		favorite = favorite || photo.PhotoFavorite
		private = private || photo.PhotoPrivate
		panorama = panorama || photo.PhotoPanorama

		if photo.PhotoTitle != "" && (title == "" || entity.SrcPriority[photo.TitleSrc] > entity.SrcPriority[titleSrc]) {
			title, titleSrc = photo.PhotoTitle, photo.TitleSrc
		}
		if photo.PhotoCaption != "" && (caption == "" || entity.SrcPriority[photo.CaptionSrc] > entity.SrcPriority[captionSrc]) {
			caption, captionSrc = photo.PhotoCaption, photo.CaptionSrc
		}
	}

	values := entity.Values{
		"photo_favorite": favorite,
		"photo_private":  private,
		"photo_panorama": panorama,
		"photo_title":    title,
		"title_src":      titleSrc,
		"photo_caption":  caption,
		"caption_src":    captionSrc,
	}
	if state.ID != canonical.ID {
		values["deleted_at"] = state.DeletedAt
		values["photo_quality"] = state.PhotoQuality
	}

	tx := entity.UnscopedDb().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	rollback := func(err error) error {
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}

	if err := tx.Model(&entity.Photo{}).Where("id = ?", canonical.ID).UpdateColumns(values).Error; err != nil {
		return rollback(err)
	}

	for _, duplicate := range duplicates {
		if err := tx.Model(&entity.File{}).Where("photo_id = ?", duplicate.ID).UpdateColumns(entity.Values{
			"photo_id":     canonical.ID,
			"photo_uid":    canonical.PhotoUID,
			"file_primary": false,
		}).Error; err != nil {
			return rollback(err)
		}

		var statements []string
		switch entity.DbDialect() {
		case dsn.DriverMySQL:
			statements = []string{
				"UPDATE IGNORE photos_keywords SET photo_id = ? WHERE photo_id = ?",
				"UPDATE IGNORE photos_labels SET photo_id = ? WHERE photo_id = ?",
				"UPDATE IGNORE photos_albums SET photo_uid = ? WHERE photo_uid = ?",
			}
		case dsn.DriverSQLite3:
			statements = []string{
				"UPDATE OR IGNORE photos_keywords SET photo_id = ? WHERE photo_id = ?",
				"UPDATE OR IGNORE photos_labels SET photo_id = ? WHERE photo_id = ?",
				"UPDATE OR IGNORE photos_albums SET photo_uid = ? WHERE photo_uid = ?",
			}
		default:
			return rollback(errors.New("unsupported database dialect"))
		}

		if err := tx.Exec(statements[0], canonical.ID, duplicate.ID).Error; err != nil {
			return rollback(err)
		} else if err = tx.Exec(statements[1], canonical.ID, duplicate.ID).Error; err != nil {
			return rollback(err)
		} else if err = tx.Exec(statements[2], canonical.PhotoUID, duplicate.PhotoUID).Error; err != nil {
			return rollback(err)
		}

		if err := tx.Exec("DELETE FROM photos_keywords WHERE photo_id = ?", duplicate.ID).Error; err != nil {
			return rollback(err)
		} else if err = tx.Exec("DELETE FROM photos_labels WHERE photo_id = ?", duplicate.ID).Error; err != nil {
			return rollback(err)
		} else if err = tx.Exec("DELETE FROM photos_albums WHERE photo_uid = ?", duplicate.PhotoUID).Error; err != nil {
			return rollback(err)
		}

		if err := tx.Model(&entity.Photo{}).Where("id = ?", duplicate.ID).UpdateColumns(entity.Values{
			"photo_quality": -1,
			"deleted_at":    entity.Now(),
		}).Error; err != nil {
			return rollback(err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	mergedUIDs := make([]string, 0, len(duplicates))
	for _, duplicate := range duplicates {
		mergedUIDs = append(mergedUIDs, duplicate.PhotoUID)
	}

	log.Infof("index: merged %s into %s for Google Pixel Camera capture %s", strings.Join(mergedUIDs, ", "), canonical.PhotoUID, related.MainLogName())

	entity.File{PhotoID: canonical.ID, PhotoUID: canonical.PhotoUID}.RegenerateIndex()
	return nil
}

// googlePixelCaptureState returns the photo whose archive state and quality a merged Google Pixel Camera capture keeps.
// That is the canonical photo unless it was removed automatically: then a visible member that existed
// when another was archived after the removal decides, else that archived one, else the canonical photo.
func googlePixelCaptureState(photos entity.Photos) *entity.Photo {
	if len(photos) == 0 || photos[0] == nil {
		return nil
	}

	canonical := photos[0]
	if !googlePixelPhotoRemoved(canonical) {
		return canonical
	}

	var archived *entity.Photo
	visible := make(entity.Photos, 0, len(photos)-1)

	for _, photo := range photos[1:] {
		switch {
		case photo == nil, googlePixelPhotoRemoved(photo):
			continue
		case photo.DeletedAt == nil:
			if photo.PhotoQuality >= 0 {
				visible = append(visible, photo)
			}
		case archived == nil && !photo.DeletedAt.Before(*canonical.DeletedAt):
			archived = photo
		}
	}

	for _, photo := range visible {
		if archived == nil || photo.CreatedAt.Before(*archived.DeletedAt) {
			return photo
		}
	}

	if archived != nil {
		return archived
	}

	return canonical
}

// googlePixelPhotoRemoved reports whether a photo was removed automatically rather than archived.
func googlePixelPhotoRemoved(photo *entity.Photo) bool {
	return photo != nil && photo.DeletedAt != nil && photo.PhotoQuality < 0
}
