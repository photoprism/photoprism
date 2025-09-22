package entity

import (
	"sync"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/pkg/rnd"
)

var photoMergeMutex = sync.Mutex{}

// ResolvePrimary ensures there is only one primary file for a photo.
func (m *Photo) ResolvePrimary() error {
	var file File

	if err := Db().Where("file_primary = TRUE AND photo_id = ?", m.ID).
		Order("file_width DESC, file_hdr DESC").
		First(&file).Error; err == nil && file.ID > 0 {
		return file.ResolvePrimary()
	}

	return nil
}

// Stackable tests if the photo may be stacked.
func (m *Photo) Stackable() bool {
	if !m.HasID() || m.PhotoStack == IsUnstacked || m.PhotoName == "" {
		return false
	}

	return true
}

// Identical returns identical photos that can be merged.
func (m *Photo) Identical(includeMeta, includeUuid bool) (identical Photos, err error) {
	if !m.Stackable() {
		return identical, nil
	}

	includeMeta = includeMeta && m.TrustedLocation() && m.TrustedTime()
	includeUuid = includeUuid && rnd.IsUUID(m.UUID)

	switch {
	case includeMeta && includeUuid:
		if err := Db().
			Where("(taken_at = ? AND taken_src = 'meta' AND place_src <> 'estimate' AND photo_stack > -1 AND cell_id = ? AND camera_serial = ? AND camera_id = ?) "+
				"OR (uuid = ? AND photo_stack > -1)"+
				"OR (photo_path = ? AND photo_name = ?)",
				m.TakenAt, m.CellID, m.CameraSerial, m.CameraID, m.UUID, m.PhotoPath, m.PhotoName).
			Order("photo_quality DESC, id ASC").Find(&identical).Error; err != nil {
			return identical, err
		}
	case includeMeta:
		if err := Db().
			Where("(taken_at = ? AND taken_src = 'meta' AND place_src <> 'estimate' AND photo_stack > -1 AND cell_id = ? AND camera_serial = ? AND camera_id = ?) "+
				"OR (photo_path = ? AND photo_name = ?)",
				m.TakenAt, m.CellID, m.CameraSerial, m.CameraID, m.PhotoPath, m.PhotoName).
			Order("photo_quality DESC, id ASC").Find(&identical).Error; err != nil {
			return identical, err
		}
	case includeUuid:
		if err := Db().
			Where("(uuid = ? AND photo_stack > -1) OR (photo_path = ? AND photo_name = ?)",
				m.UUID, m.PhotoPath, m.PhotoName).
			Order("photo_quality DESC, id ASC").Find(&identical).Error; err != nil {
			return identical, err
		}
	default:
		if err := Db().
			Where("photo_path = ? AND photo_name = ?", m.PhotoPath, m.PhotoName).
			Order("photo_quality DESC, id ASC").Find(&identical).Error; err != nil {
			return identical, err
		}
	}

	return identical, nil
}

// Merge photo with identical ones.
func (m *Photo) Merge(mergeMeta, mergeUuid bool) (original Photo, merged Photos, err error) {
	photoMergeMutex.Lock()
	defer photoMergeMutex.Unlock()

	identical, err := m.Identical(mergeMeta, mergeUuid)

	if len(identical) < 2 || err != nil {
		return Photo{}, merged, err
	}

	logResult := func(res *gorm.DB) {
		if res.Error != nil {
			log.Errorf("merge: %s", res.Error.Error())
			err = res.Error
		}
	}

	for i, merge := range identical {
		if i == 0 {
			original = *merge
			log.Debugf("photo: merging id %d with %d identical", original.ID, len(identical)-1)
			continue
		}

		deleted := gorm.DeletedAt{Time: Now(), Valid: true}

		logResult(UnscopedDb().Exec("UPDATE files SET photo_id = ?, photo_uid = ?, file_primary = FALSE WHERE photo_id = ?", original.ID, original.PhotoUID, merge.ID))
		logResult(UnscopedDb().Exec("UPDATE photos SET photo_quality = -1, deleted_at = ? WHERE id = ?", Now(), merge.ID))

		switch DbDialect() {
		case Postgres:
			logResult(UnscopedDb().Exec("UPDATE photos_keywords SET photo_id = ? WHERE photo_id = ? AND keyword_id NOT IN (SELECT keyword_id FROM photos_keywords WHERE photo_id = ?)", original.ID, merge.ID, original.ID))
			logResult(UnscopedDb().Exec("UPDATE photos_labels SET photo_id = ? WHERE photo_id = ? AND label_id NOT IN (SELECT label_id FROM photos_labels WHERE photo_id = ?)", original.ID, merge.ID, original.ID))
			logResult(UnscopedDb().Exec("UPDATE photos_albums SET photo_uid = ? WHERE photo_uid = ? AND album_uid NOT IN (SELECT album_uid FROM photos_albums WHERE photo_uid = ?)", original.PhotoUID, merge.PhotoUID, original.PhotoUID))
		case MySQL:
			logResult(UnscopedDb().Exec("UPDATE IGNORE photos_keywords SET photo_id = ? WHERE photo_id = ?", original.ID, merge.ID))
			logResult(UnscopedDb().Exec("UPDATE IGNORE photos_labels SET photo_id = ? WHERE photo_id = ?", original.ID, merge.ID))
			logResult(UnscopedDb().Exec("UPDATE IGNORE photos_albums SET photo_uid = ? WHERE photo_uid = ?", original.PhotoUID, merge.PhotoUID))
		case SQLite3:
			logResult(UnscopedDb().Exec("UPDATE OR IGNORE photos_keywords SET photo_id = ? WHERE photo_id = ?", original.ID, merge.ID))
			logResult(UnscopedDb().Exec("UPDATE OR IGNORE photos_labels SET photo_id = ? WHERE photo_id = ?", original.ID, merge.ID))
			logResult(UnscopedDb().Exec("UPDATE OR IGNORE photos_albums SET photo_uid = ? WHERE photo_uid = ?", original.PhotoUID, merge.PhotoUID))
		default:
			log.Warnf("sql: unsupported dialect %s", DbDialect())
		}

		merge.DeletedAt = deleted
		merge.PhotoQuality = -1

		merged = append(merged, merge)
	}

	if original.ID != m.ID {
		deleted := gorm.DeletedAt{Time: Now(), Valid: true}
		m.DeletedAt = deleted
		m.PhotoQuality = -1
	}

	File{PhotoID: original.ID, PhotoUID: original.PhotoUID}.RegenerateIndex()

	return original, merged, err
}
