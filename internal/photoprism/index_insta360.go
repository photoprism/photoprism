package photoprism

import (
	"errors"
	"sync"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
)

var insta360ReconcileMutex sync.Mutex

// reconcileInsta360Photos combines previously separate capture records during a forced reindex.
func reconcileInsta360Photos(related RelatedFiles) error {
	if related.Main == nil {
		return nil
	}

	capture := FindInsta360Capture(related.Main)
	if capture == nil || !capture.ValidPair() {
		return nil
	}

	fileNames := make([]string, 0, 3)
	for _, file := range capture.Files() {
		fileNames = append(fileNames, file.RootRelName())
	}

	insta360ReconcileMutex.Lock()
	defer insta360ReconcileMutex.Unlock()

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

	canonical := photos[0]
	favorite, private, panorama, allArchived := false, false, false, true
	title, titleSrc := canonical.PhotoTitle, canonical.TitleSrc
	caption, captionSrc := canonical.PhotoCaption, canonical.CaptionSrc

	for _, photo := range photos {
		favorite = favorite || photo.PhotoFavorite
		private = private || photo.PhotoPrivate
		panorama = panorama || photo.PhotoPanorama
		allArchived = allArchived && photo.DeletedAt != nil

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
	if allArchived {
		values["deleted_at"] = canonical.DeletedAt
	} else {
		values["deleted_at"] = nil
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

	for _, duplicate := range photos[1:] {
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

	entity.File{PhotoID: canonical.ID, PhotoUID: canonical.PhotoUID}.RegenerateIndex()
	return nil
}
