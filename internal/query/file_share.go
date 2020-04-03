package query

import "github.com/photoprism/photoprism/internal/entity"

// FileShares
func (q *Query) FileShares(accountId uint, status string) (result []entity.FileShare, err error) {
	s := q.db.Where(&entity.FileShare{})

	if accountId > 0 {
		s = s.Where("account_id = ?", accountId)
	}

	if status != "" {
		s = s.Where("status = ?", status)
	}

	s = s.Order("created_at ASC")
	s = s.Limit(100).Offset(0)

	s = s.Preload("File")

	if err := s.Find(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}
