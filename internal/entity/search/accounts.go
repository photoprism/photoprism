package search

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// Accounts returns a list of accounts.
func Accounts(f form.SearchServices) (result entity.Services, err error) {
	s := Db().Where(&entity.Service{})

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
