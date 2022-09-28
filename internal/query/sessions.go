package query

import (
	"time"

	"github.com/photoprism/photoprism/internal/entity"
)

// Sessions returns stored sessions.
func Sessions() (result entity.Sessions, err error) {
	err = Db().
		Table(entity.Session{}.TableName()).
		Select("*").
		Where("expires_at > ?", time.Now()).
		Scan(&result).Error

	return result, err
}

// Session finds an existing session by id.
func Session(id string) (result entity.Session, err error) {
	err = Db().Where("id = ?", id).First(&result).Error

	return result, err
}
