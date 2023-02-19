package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// FileSyncs returns a list of FileSync entities for a given account and status.
func FileSyncs(accountId uint, status string, limit int) (result []entity.FileSync, err error) {
	s := Db().Where(&entity.FileSync{})

	if accountId > 0 {
		s = s.Where("service_id = ?", accountId)
	}

	if status != "" {
		s = s.Where("status = ?", status)
	}

	s = s.Order("remote_name ASC")

	if limit > 0 {
		s = s.Limit(limit).Offset(0)
	}

	s = s.Preload("File")

	if err := s.Find(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}
