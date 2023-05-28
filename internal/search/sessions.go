package search

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Sessions finds user sessions.
func Sessions(f form.SearchSessions) (result entity.Sessions, err error) {
	result = entity.Sessions{}
	stmt := Db()

	search := strings.TrimSpace(f.Query)
	uid := strings.TrimSpace(f.UID)
	sortOrder := f.Order
	limit := f.Count
	offset := f.Offset

	if search == "all" {
		// Don't filter.
	} else if rnd.IsUID(uid, entity.UserUID) {
		stmt = stmt.Where("user_uid = ?", search)
	} else if search != "" {
		stmt = stmt.Where("user_name LIKE ? OR auth_provider LIKE ?", search+"%", search+"%")
	}

	if sortOrder == "" {
		sortOrder = "last_active DESC, user_name"
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
