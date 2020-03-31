package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// Accounts returns a list of accounts.
func (q *Query) Accounts(f form.AccountSearch) (result []entity.Account, err error) {
	s := q.db.Where(&entity.Account{})

	if f.Share {
		s = s.Where("acc_share = 1")
	}

	if f.Sync {
		s = s.Where("acc_sync = 1")
	}

	s = s.Order("acc_name ASC")

	if f.Count > 0 && f.Count <= 1000 {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(1000).Offset(0)
	}

	if err := s.Find(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}

// AccountByID finds an account by primary key.
func (q *Query) AccountByID(id uint) (result entity.Account, err error) {
	if err := q.db.Where("id = ?", id).First(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}
