package query

import (
	"errors"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
)

// Session finds an existing session by its id.
func Session(id string) (result entity.Session, err error) {
	if l := len(id); l < 6 || l > 2048 {
		return result, errors.New("invalid session id")
	} else if rnd.IsRefID(id) {
		err = Db().Where("ref_id = ?", id).First(&result).Error
	} else if rnd.IsSessionID(id) {
		err = Db().Where("id LIKE ?", id).First(&result).Error
	} else {
		err = Db().Where("id LIKE ?", rnd.SessionID(id)).First(&result).Error
	}

	return result, err
}

// Sessions finds user sessions and returns them.
func Sessions(limit, offset int, sortOrder, search string) (result entity.Sessions, err error) {
	result = entity.Sessions{}
	stmt := Db()

	search = strings.TrimSpace(search)

	if search == "expired" {
		stmt = stmt.Where("sess_expires > 0 AND sess_expires < ?", unix.Now())
	} else if rnd.IsSessionID(search) {
		stmt = stmt.Where("id = ?", search)
	} else if rnd.IsAuthToken(search) {
		stmt = stmt.Where("id = ?", rnd.SessionID(search))
	} else if rnd.IsUID(search, entity.UserUID) {
		stmt = stmt.Where("user_uid = ?", search)
	} else if search != "" {
		stmt = stmt.Where("user_name LIKE ? OR auth_provider LIKE ?", search+"%", search+"%")
	}

	if sortOrder == "" {
		sortOrder = "last_active DESC, created_at DESC, user_name"
	}

	if limit > 0 {
		stmt = stmt.Limit(limit)

		if offset > 0 {
			stmt = stmt.Offset(offset)
		}
	}

	err = stmt.Order(sortOrder).Find(&result).Error

	return result, err
}
