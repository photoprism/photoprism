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
	if err := UnscopedDb().Delete(entity.Duplicate{}, "file_hash IN (SELECT d.file_hash FROM duplicates d LEFT JOIN files f ON d.file_hash = f.file_hash AND f.file_missing = 0 AND f.deleted_at IS NULL WHERE f.file_hash IS NULL)").Error; err == nil {
		return nil
	}

	// MySQL fallback, see https://github.com/photoprism/photoprism/issues/599
	return UnscopedDb().Delete(entity.Duplicate{}, "file_hash IN (SELECT file_hash FROM (SELECT d.file_hash FROM duplicates d LEFT JOIN files f ON d.file_hash = f.file_hash AND f.file_missing = 0 AND f.deleted_at IS NULL WHERE f.file_hash IS NULL) AS tmp)").Error
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
