package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// FindAccounts returns a list of accounts.
func (s *Query) Accounts(f form.AccountSearch) (result []entity.Account, err error) {
	if err := s.db.Where(&entity.Account{}).Limit(f.Count).Offset(f.Offset).Find(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}

// AccountByID finds an account by primary key.
func (s *Query) AccountByID(id uint) (result entity.Account, err error) {
	if err := s.db.Where("id = ?", id).First(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}
