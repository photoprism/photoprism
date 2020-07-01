package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// AccountSearch returns a list of accounts.
func AccountSearch(f form.AccountSearch) (result entity.Accounts, err error) {
	s := Db().Where(&entity.Account{})

	if f.Share {
		s = s.Where("acc_share = 1")
	}

	if f.Sync {
		s = s.Where("acc_sync = 1")
	}

	if f.Status != "" {
		s = s.Where("sync_status = ?", f.Status)
	}

	s = s.Order("acc_name ASC")

	if f.Count > 0 && f.Count <= MaxResults {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(f.Offset)
	}

	if err := s.Find(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}

// AccountByID finds an account by primary key.
func AccountByID(id uint) (result entity.Account, err error) {
	if err := Db().Where("id = ?", id).First(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}
