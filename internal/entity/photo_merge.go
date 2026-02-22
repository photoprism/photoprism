package entity

import (
	"sync"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var photoMergeMutex = sync.Mutex{}

// ResolvePrimary ensures only one associated file remains marked as primary, delegating to the file helper.
func (m *Photo) ResolvePrimary() error {
	var file File

	if err := Db().Where("file_primary = 1 AND photo_id = ?", m.ID).
		Order("file_width DESC, file_hdr DESC").
		First(&file).Error; err == nil && file.ID > 0 {
		return file.ResolvePrimary()
	}

	return nil
}

// Stackable reports whether the photo participates in stacking workflows.
func (m *Photo) Stackable() bool {
	if !m.HasID() || m.PhotoStack == IsUnstacked || m.PhotoName == "" {
		return false
	}

	return true
}

// Identical returns candidate photos that can be merged with the current one based on metadata and/or UUID.
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

// Merge collapses identical photos into a single original, reassigning files and associations while marking duplicates deleted.
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

		deleted := Now()

		logResult(UnscopedDb().Exec("UPDATE files SET photo_id = ?, photo_uid = ?, file_primary = 0 WHERE photo_id = ?", original.ID, original.PhotoUID, merge.ID))
		logResult(UnscopedDb().Exec("UPDATE photos SET photo_quality = -1, deleted_at = ? WHERE id = ?", Now(), merge.ID))

		switch DbDialect() {
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

		merge.DeletedAt = &deleted
		merge.PhotoQuality = -1

		merged = append(merged, merge)
	}

	if updateErr := original.SyncMediaTypeFromFiles(SrcFile); updateErr != nil {
		log.Warnf("merge: %s while syncing media type of %s", updateErr, original.String())
	}

	if original.ID != m.ID {
		deleted := Now()
		m.DeletedAt = &deleted
		m.PhotoQuality = -1
	} else {
		m.PhotoType = original.PhotoType
		m.TypeSrc = original.TypeSrc
	}

	File{PhotoID: original.ID, PhotoUID: original.PhotoUID}.RegenerateIndex()

	return original, merged, err
}

// SyncMediaTypeFromFiles updates PhotoType to the highest-priority media type of active non-sidecar files.
func (m *Photo) SyncMediaTypeFromFiles(typeSrc string) error {
	if m == nil || !m.HasID() {
		return nil
	}

	// Keep explicit user/admin overrides untouched.
	if SrcPriority[m.TypeSrc] > SrcPriority[typeSrc] {
		return nil
	}

	var mediaTypes []string

	if err := UnscopedDb().
		Model(File{}).
		Where("photo_id = ? AND file_missing = 0 AND file_sidecar = 0 AND deleted_at IS NULL", m.ID).
		Pluck("media_type", &mediaTypes).Error; err != nil {
		return err
	}

	bestType := media.Image
	found := false

	for _, mediaType := range mediaTypes {
		t := media.Type(mediaType)

		if !t.IsMain() {
			continue
		}

		if !found || media.Priority[t] > media.Priority[bestType] {
			bestType = t
			found = true
		}
	}

	// Do not change the current type when no eligible main media file exists.
	if !found {
		return nil
	}

	previousType := m.MediaType()
	m.SetMediaType(bestType, typeSrc)

	if m.MediaType() == previousType {
		return nil
	}

	return m.Updates(Values{
		"PhotoType": m.PhotoType,
		"TypeSrc":   m.TypeSrc,
	})
}
