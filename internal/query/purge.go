package query

import "github.com/photoprism/photoprism/internal/entity"

// PurgeOrphans removes orphan database entries.
func PurgeOrphans() error {
	if err := PurgeOrphanDuplicates(); err != nil {
		return err
	}
	if err := PurgeOrphanCountries(); err != nil {
		return err
	}
	if err := PurgeOrphanCameras(); err != nil {
		return err
	}
	if err := PurgeOrphanLenses(); err != nil {
		return err
	}

	return nil
}

// PurgeOrphanDuplicates deletes all files from the duplicates table that don't exist in the files table.
func PurgeOrphanDuplicates() error {
	return UnscopedDb().Delete(
		entity.Duplicate{},
		"file_hash NOT IN (SELECT file_hash FROM files WHERE file_missing = 0 AND deleted_at IS NULL)").Error
}

// PurgeOrphanCountries removes countries without any photos.
func PurgeOrphanCountries() error {
	entity.FlushCountryCache()
	switch DbDialect() {
	default:
		return UnscopedDb().Exec(`DELETE FROM countries WHERE country_slug <> ? AND id NOT IN (SELECT photo_country FROM photos)`, entity.UnknownCountry.CountrySlug).Error
	}
}

// PurgeOrphanCameras removes cameras without any photos.
func PurgeOrphanCameras() error {
	entity.FlushCameraCache()
	switch DbDialect() {
	default:
		return UnscopedDb().Exec(`DELETE FROM cameras WHERE camera_slug <> ? AND id NOT IN (SELECT camera_id FROM photos)`, entity.UnknownCamera.CameraSlug).Error
	}
}

// PurgeOrphanLenses removes cameras without any photos.
func PurgeOrphanLenses() error {
	entity.FlushLensCache()
	switch DbDialect() {
	default:
		return UnscopedDb().Exec(`DELETE FROM lenses WHERE lens_slug <> ? AND id NOT IN (SELECT lens_id FROM photos)`, entity.UnknownLens.LensSlug).Error
	}
}
