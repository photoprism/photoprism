package query

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
)

// Files returns file entities in the range of limit and offset sorted by id.
func (q *Query) Files(limit int, offset int, filePath string) (files []entity.File, err error) {
	if strings.HasPrefix(filePath, "/") {
		filePath = filePath[1:]
	}

	stmt := q.db.Unscoped()

	if filePath != "" {
		stmt = stmt.Where("file_name LIKE ?", filePath + "/%")
	}

	err = stmt.Order("id").Limit(limit).Offset(offset).Find(&files).Error

	return files, err
}

// FilesByUUID
func (q *Query) FilesByUUID(u []string, limit int, offset int) (files []entity.File, err error) {
	if err := q.db.Where("(photo_uuid IN (?) AND file_primary = 1) OR file_uuid IN (?)", u, u).Preload("Photo").Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		return files, err
	}

	return files, nil
}

// FileByPhotoUUID
func (q *Query) FileByPhotoUUID(u string) (file entity.File, err error) {
	if err := q.db.Where("photo_uuid = ? AND file_primary = 1", u).Preload("Links").Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FileByUUID returns the file entity for a given UUID.
func (q *Query) FileByUUID(uuid string) (file entity.File, err error) {
	if err := q.db.Where("file_uuid = ?", uuid).Preload("Links").Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FirstFileByHash finds a file with a given hash string.
func (q *Query) FileByHash(fileHash string) (file entity.File, err error) {
	if err := q.db.Where("file_hash = ?", fileHash).Preload("Links").Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// SetPhotoPrimary sets a new primary image file for a photo.
func (q *Query) SetPhotoPrimary(photoUUID, fileUUID string) error {
	q.db.Model(entity.File{}).Where("photo_uuid = ? AND file_uuid <> ?", photoUUID, fileUUID).UpdateColumn("file_primary", false)
	return q.db.Model(entity.File{}).Where("photo_uuid = ? AND file_uuid = ?", photoUUID, fileUUID).UpdateColumn("file_primary", true).Error
}
