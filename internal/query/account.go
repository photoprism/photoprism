package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// AccountByID finds an account by primary key.
func AccountByID(id uint) (result entity.Account, err error) {
	err = Db().Where("id = ?", id).First(&result).Error

	return result, err
}
