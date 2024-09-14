package search

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Users finds registered users.
func Users(f form.SearchUsers) (result entity.Users, err error) {
	result = entity.Users{}
	stmt := Db()

	search := strings.TrimSpace(f.Query)
	sortOrder := f.Order
	limit := f.Count
	offset := f.Offset

	if search == "all" {
		// Don't filter.
	} else if id := txt.Int(search); id != 0 {
		stmt = stmt.Where("id = ?", id)
	} else if rnd.IsUID(search, entity.UserUID) {
		stmt = stmt.Where("user_uid = ?", search)
	} else if search != "" {
		stmt = stmt.Where("user_name LIKE ? OR user_email LIKE ? OR display_name LIKE ?", search+"%", search+"%", search+"%")
	} else {
		stmt = stmt.Where("id > 0")
	}

	if sortOrder == "" {
		sortOrder = "user_name, id"
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
