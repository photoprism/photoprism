package query

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
)

// ExistingFiles returns not-missing and not-deleted file entities in the range of limit and offset sorted by id.
func ExistingFiles(limit int, offset int, filePath string) (files []entity.File, err error) {
	if strings.HasPrefix(filePath, "/") {
		filePath = filePath[1:]
	}

	stmt := Db().Unscoped().Where("file_missing = 0 AND deleted_at IS NULL")

	if filePath != "" {
		stmt = stmt.Where("file_name LIKE ?", filePath+"/%")
	}

	err = stmt.Order("id").Limit(limit).Offset(offset).Find(&files).Error

	return files, err
}

// FilesByUID
func FilesByUID(u []string, limit int, offset int) (files []entity.File, err error) {
	if err := Db().Where("(photo_uid IN (?) AND file_primary = 1) OR file_uid IN (?)", u, u).Preload("Photo").Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		return files, err
	}

	return files, nil
}

// FileByPhotoUID
func FileByPhotoUID(u string) (file entity.File, err error) {
	if err := Db().Where("photo_uid = ? AND file_primary = 1", u).Preload("Links").Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// VideoByPhotoUID
func VideoByPhotoUID(u string) (file entity.File, err error) {
	if err := Db().Where("photo_uid = ? AND file_video = 1", u).Preload("Links").Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FileByUID returns the file entity for a given UID.
func FileByUID(uid string) (file entity.File, err error) {
	if err := Db().Where("file_uid = ?", uid).Preload("Links").Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FileByHash finds a file with a given hash string.
func FileByHash(fileHash string) (file entity.File, err error) {
	if err := Db().Where("file_hash = ?", fileHash).Preload("Links").Preload("Photo").First(&file).Error; err != nil {
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
