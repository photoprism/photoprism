package query

import "github.com/photoprism/photoprism/internal/entity"

// FindFiles finds files returning maximum results defined by limit
// and finding them from an offest defined by offset.
func (s *Repo) FindFiles(limit int, offset int) (files []entity.File, err error) {
	if err := s.db.Where(&entity.File{}).Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		return files, err
	}

	return files, nil
}

// FindFilesByUUID
func (s *Repo) FindFilesByUUID(u []string, limit int, offset int) (files []entity.File, err error) {
	if err := s.db.Where("(photo_uuid IN (?) AND file_primary = 1) OR file_uuid IN (?)", u, u).Preload("Photo").Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		return files, err
	}

	return files, nil
}

// FindFileByPhotoUUID
func (s *Repo) FindFileByPhotoUUID(u string) (file entity.File, err error) {
	if err := s.db.Where("photo_uuid = ? AND file_primary = 1", u).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FindFileByID returns a MediaFile given a certain ID.
func (s *Repo) FindFileByID(id string) (file entity.File, err error) {
	if err := s.db.Where("id = ?", id).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// FindFileByHash finds a file with a given hash string.
func (s *Repo) FindFileByHash(fileHash string) (file entity.File, err error) {
	if err := s.db.Where("file_hash = ?", fileHash).Preload("Photo").First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}
