package search

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// Sessions finds user sessions.
func Sessions(f form.SearchSessions) (result entity.Sessions, err error) {
	result = entity.Sessions{}
	stmt := Db()

	userUid := strings.TrimSpace(f.UID)
	search := strings.TrimSpace(f.Query)

	sortOrder := f.Order
	limit := f.Count
	offset := f.Offset

	// Filter by user UID?
	if userUid != "" {
		stmt = stmt.Where("user_uid = ?", userUid)
	}

	// Filter by user name and/or auth provider name?
	if search != "" && search != "all" {
		stmt = stmt.Where("user_name LIKE ? OR auth_provider LIKE ?", search+"%", search+"%")
	}

	// Sort results?
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
