package query

import (
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
)

// PurgeOrphans removes orphan database entries.
func PurgeOrphans() error {
	// Remove files without a photo.
	start := time.Now()
	if count, err := PurgeOrphanFiles(); err != nil {
		return err
	} else if count > 0 {
		log.Infof("index: removed %d orphan files [%s]", count, time.Since(start))
	} else {
		log.Debugf("index: found no orphan files [%s]", time.Since(start))
	}

	// Remove duplicates without an original file.
	if err := PurgeOrphanDuplicates(); err != nil {
		return err
	}

	// Remove unused countries.
	if err := PurgeOrphanCountries(); err != nil {
		return err
	}

	// Reset references to cameras and lenses that no longer exist.
	if count, err := ResetMissingCameras(); err != nil {
		return err
	} else if count > 0 {
		log.Infof("index: reset %s with a missing camera", english.Plural(int(count), "picture", "pictures"))
	}

	if count, err := ResetMissingLenses(); err != nil {
		return err
	} else if count > 0 {
		log.Infof("index: reset %s with a missing lens", english.Plural(int(count), "picture", "pictures"))
	}

	// Remove unused cameras.
	if err := PurgeOrphanCameras(); err != nil {
		return err
	}

	// Remove unused lenses.
	return PurgeOrphanLenses()
}

// PurgeOrphanFiles removes files without a photo from the index.
func PurgeOrphanFiles() (count int, err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	files, err := OrphanFiles()

	if err != nil {
		return count, err
	}

	for i := range files {
		if err = files[i].DeletePermanently(); err != nil {
			return count, err
		}

		count++
	}

	return count, err
}

// PurgeOrphanDuplicates deletes all files from the duplicates table that don't exist in the files table.
func PurgeOrphanDuplicates() error {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	result := UnscopedDb().
		Delete(entity.Duplicate{},
			"file_hash NOT IN (SELECT file_hash FROM files WHERE file_missing = 0 AND deleted_at IS NULL)")

	return result.Error
}

// PurgeOrphanCountries removes countries without any photos.
func PurgeOrphanCountries() error {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	entity.FlushCountryCache()

	result := UnscopedDb().
		Exec(`DELETE FROM countries WHERE country_slug <> ? AND id NOT IN (SELECT photo_country FROM photos)`,
			entity.UnknownCountry.CountrySlug)

	return result.Error
}

// PurgeOrphanCameras removes cameras without any photos, except those that have been added manually.
func PurgeOrphanCameras() error {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	entity.FlushCameraCache()

	result := UnscopedDb().
		Exec(`DELETE FROM cameras WHERE camera_slug <> ? AND (camera_src IS NULL OR camera_src <> ?) AND id NOT IN (SELECT camera_id FROM photos)`,
			entity.UnknownCamera.CameraSlug, entity.SrcManual)

	return result.Error
}

// PurgeOrphanLenses removes lenses without any photos, except those that have been added manually.
func PurgeOrphanLenses() error {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	entity.FlushLensCache()

	result := UnscopedDb().
		Exec(`DELETE FROM lenses WHERE lens_slug <> ? AND (lens_src IS NULL OR lens_src <> ?) AND id NOT IN (SELECT lens_id FROM photos)`,
			entity.UnknownLens.LensSlug, entity.SrcManual)

	return result.Error
}

// ResetMissingCameras assigns pictures that reference a camera which no longer exists to the unknown camera.
// Such references remain e.g. when a camera is deleted with the CLI while another process still has it cached.
func ResetMissingCameras() (int64, error) {
	if entity.UnknownCamera.ID == 0 {
		return 0, nil
	}

	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	// Find missing IDs first, so that no rows are locked for writing unless there is something to repair.
	var missing []uint

	if err := UnscopedDb().Raw(`SELECT DISTINCT camera_id FROM photos WHERE camera_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM cameras WHERE cameras.id = photos.camera_id)`).
		Pluck("camera_id", &missing).Error; err != nil {
		return 0, err
	} else if len(missing) == 0 {
		return 0, nil
	}

	res := UnscopedDb().Model(&entity.Photo{}).Where("camera_id IN (?)", missing).UpdateColumn("camera_id", entity.UnknownCamera.ID)

	return res.RowsAffected, res.Error
}

// ResetMissingLenses assigns pictures that reference a lens which no longer exists to the unknown lens.
// Such references remain e.g. when a lens is deleted with the CLI while another process still has it cached.
func ResetMissingLenses() (int64, error) {
	if entity.UnknownLens.ID == 0 {
		return 0, nil
	}

	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	// Find missing IDs first, so that no rows are locked for writing unless there is something to repair.
	var missing []uint

	if err := UnscopedDb().Raw(`SELECT DISTINCT lens_id FROM photos WHERE lens_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM lenses WHERE lenses.id = photos.lens_id)`).
		Pluck("lens_id", &missing).Error; err != nil {
		return 0, err
	} else if len(missing) == 0 {
		return 0, nil
	}

	res := UnscopedDb().Model(&entity.Photo{}).Where("lens_id IN (?)", missing).UpdateColumn("lens_id", entity.UnknownLens.ID)

	return res.RowsAffected, res.Error
}
