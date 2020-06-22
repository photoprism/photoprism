package query

import (
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
