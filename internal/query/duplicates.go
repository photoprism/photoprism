package query

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
)

// CleanDuplicates removes all files from the duplicates table that don't exist in the files table.
func CleanDuplicates() error {
	if err := UnscopedDb().Delete(entity.Duplicate{}, "file_hash IN (SELECT d.file_hash FROM duplicates d LEFT JOIN files f ON d.file_hash = f.file_hash AND f.file_missing = 0 AND f.deleted_at IS NULL WHERE f.file_hash IS NULL)").Error; err == nil {
		return nil
	}

	// MySQL fallback, see https://github.com/photoprism/photoprism/issues/599
	return UnscopedDb().Delete(entity.Duplicate{}, "file_hash IN (SELECT file_hash FROM (SELECT d.file_hash FROM duplicates d LEFT JOIN files f ON d.file_hash = f.file_hash AND f.file_missing = 0 AND f.deleted_at IS NULL WHERE f.file_hash IS NULL) AS tmp)").Error
}

// Duplicates returns duplicate files in the range of limit and offset sorted by file name.
func Duplicates(limit, offset int, pathName string) (files entity.Duplicates, err error) {
	if strings.HasPrefix(pathName, "/") {
		pathName = pathName[1:]
	}

	stmt := Db()

	if pathName != "" {
		stmt = stmt.Where("file_name LIKE ?", pathName+"/%")
	}

	err = stmt.Order("file_name").Limit(limit).Offset(offset).Find(&files).Error

	return files, err
}
