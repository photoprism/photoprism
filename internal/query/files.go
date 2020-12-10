package query

import (
	"path"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
)

// FilesByPath returns a slice of files in a given originals folder.
func FilesByPath(limit, offset int, rootName, pathName string) (files entity.Files, err error) {
	if strings.HasPrefix(pathName, "/") {
		pathName = pathName[1:]
	}

	err = Db().
		Table("files").Select("files.*").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.deleted_at IS NULL").
		Where("files.file_missing = 0").
		Where("files.file_root = ? AND photos.photo_path = ?", rootName, pathName).
		Order("files.file_name").
		Limit(limit).Offset(offset).
		Find(&files).Error

	return files, err
}

// Files returns not-missing and not-deleted file entities in the range of limit and offset sorted by id.
func Files(limit, offset int, pathName string, includeMissing bool) (files entity.Files, err error) {
	if strings.HasPrefix(pathName, "/") {
		pathName = pathName[1:]
	}

	stmt := Db()

	if !includeMissing {
		stmt = stmt.Where("file_missing = 0")
	}

	if pathName != "" {
		stmt = stmt.Where("files.file_name LIKE ?", pathName+"/%")
	}

	err = stmt.Order("id").Limit(limit).Offset(offset).Find(&files).Error

	return files, err
}

// FilesByUID
func FilesByUID(u []string, limit int, offset int) (files entity.Files, err error) {
	if err := Db().Where("(photo_uid IN (?) AND file_primary = 1) OR file_uid IN (?)", u, u).Preload("Photo").Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		return files, err
	}

	return files, nil
}

// FileByPhotoUID
func FileByPhotoUID(u string) (file entity.File, err error) {
	if err := Db().Where("photo_uid = ? AND file_primary = 1", u).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// VideoByPhotoUID
func VideoByPhotoUID(u string) (file entity.File, err error) {
	if err := Db().Where("photo_uid = ? AND file_video = 1", u).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FileByUID returns the file entity for a given UID.
func FileByUID(uid string) (file entity.File, err error) {
	if err := Db().Where("file_uid = ?", uid).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FileByHash finds a file with a given hash string.
func FileByHash(fileHash string) (file entity.File, err error) {
	if err := Db().Where("file_hash = ?", fileHash).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// SetPhotoPrimary sets a new primary image file for a photo.
func SetPhotoPrimary(photoUID, fileUID string) error {
	Db().Model(entity.File{}).Where("photo_uid = ? AND file_uid <> ?", photoUID, fileUID).UpdateColumn("file_primary", false)
	return Db().Model(entity.File{}).Where("photo_uid = ? AND file_uid = ?", photoUID, fileUID).UpdateColumn("file_primary", true).Error
}

// SetFileError updates the file error column.
func SetFileError(fileUID, errorString string) {
	if err := Db().Model(entity.File{}).Where("file_uid = ?", fileUID).UpdateColumn("file_error", errorString).Error; err != nil {
		log.Errorf("query: %s", err.Error())
	}
}

type FileMap map[string]int64

// IndexedFiles returns a map of already indexed files with their mod time unix timestamp as value.
func IndexedFiles() (result FileMap, err error) {
	result = make(FileMap)

	type File struct {
		FileRoot string
		FileName string
		ModTime  int64
	}

	// Query known duplicates.
	var duplicates []File

	if err := UnscopedDb().Raw("SELECT file_root, file_name, mod_time FROM duplicates").Scan(&duplicates).Error; err != nil {
		return result, err
	}

	for _, row := range duplicates {
		result[path.Join(row.FileRoot, row.FileName)] = row.ModTime
	}

	// Query indexed files.
	var files []File

	if err := UnscopedDb().Raw("SELECT file_root, file_name, mod_time FROM files WHERE file_missing = 0 AND deleted_at IS NULL").Scan(&files).Error; err != nil {
		return result, err
	}

	for _, row := range files {
		result[path.Join(row.FileRoot, row.FileName)] = row.ModTime
	}

	return result, err
}

// CleanDuplicates removes all files from the duplicates table that don't exist in the files table.
func CleanDuplicates() error {
	if err := UnscopedDb().Delete(entity.Duplicate{}, "file_hash IN (SELECT d.file_hash FROM duplicates d LEFT JOIN files f ON d.file_hash = f.file_hash AND f.file_missing = 0 AND f.deleted_at IS NULL WHERE f.file_hash IS NULL)").Error; err == nil {
		return nil
	}

	// MySQL fallback, see https://github.com/photoprism/photoprism/issues/599
	return UnscopedDb().Delete(entity.Duplicate{}, "file_hash IN (SELECT file_hash FROM (SELECT d.file_hash FROM duplicates d LEFT JOIN files f ON d.file_hash = f.file_hash AND f.file_missing = 0 AND f.deleted_at IS NULL WHERE f.file_hash IS NULL) AS tmp)").Error
}
