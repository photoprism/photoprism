package query

import "github.com/photoprism/photoprism/internal/entity"

// Files finds files returning maximum results defined by limit
// and finding them from an offest defined by offset.
func (s *Query) Files(limit int, offset int) (files []entity.File, err error) {
	if err := s.db.Where(&entity.File{}).Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		return files, err
	}

	return files, nil
}

// FilesByUUID
func (s *Query) FilesByUUID(u []string, limit int, offset int) (files []entity.File, err error) {
	if err := s.db.Where("(photo_uuid IN (?) AND file_primary = 1) OR file_uuid IN (?)", u, u).Preload("Photo").Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		return files, err
	}

	return files, nil
}

// FileByPhotoUUID
func (s *Query) FileByPhotoUUID(u string) (file entity.File, err error) {
	if err := s.db.Where("photo_uuid = ? AND file_primary = 1", u).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FileByID returns a MediaFile given a certain ID.
func (s *Query) FileByID(id string) (file entity.File, err error) {
	if err := s.db.Where("id = ?", id).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FirstFileByHash finds a file with a given hash string.
func (s *Query) FileByHash(fileHash string) (file entity.File, err error) {
	if err := s.db.Where("file_hash = ?", fileHash).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}
