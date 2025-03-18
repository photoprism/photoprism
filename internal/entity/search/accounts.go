package search

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// Accounts returns a list of accounts.
func Accounts(frm form.SearchServices) (result entity.Services, err error) {
	s := Db().Where(&entity.Service{})

	if frm.Share {
		s = s.Where("acc_share = 1")
	}

	if frm.Sync {
		s = s.Where("acc_sync = 1")
	}

	if frm.Status != "" {
		s = s.Where("sync_status = ?", frm.Status)
	}

	s = s.Order("acc_name ASC")

	if frm.Count > 0 && frm.Count <= MaxResults {
		s = s.Limit(frm.Count).Offset(frm.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(frm.Offset)
	}

	if err := s.Find(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}
