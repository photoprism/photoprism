package query

import (
	"time"

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

	// Remove unused cameras.
	if err := PurgeOrphanCameras(); err != nil {
		return err
	}

	// Remove unused camera lenses.
	if err := PurgeOrphanLenses(); err != nil {
		return err
	}

	return nil
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

// PurgeOrphanCameras removes cameras without any photos.
func PurgeOrphanCameras() error {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	entity.FlushCameraCache()

	result := UnscopedDb().
		Exec(`DELETE FROM cameras WHERE camera_slug <> ? AND id NOT IN (SELECT camera_id FROM photos)`,
			entity.UnknownCamera.CameraSlug)

	return result.Error
}

// PurgeOrphanLenses removes cameras without any photos.
func PurgeOrphanLenses() error {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	entity.FlushLensCache()

	result := UnscopedDb().
		Exec(`DELETE FROM lenses WHERE lens_slug <> ? AND id NOT IN (SELECT lens_id FROM photos)`,
			entity.UnknownLens.LensSlug)

	return result.Error
}
