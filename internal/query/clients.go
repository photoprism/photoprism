package query

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Clients finds clients and returns them.
func Clients(limit, offset int, sortOrder, search string) (result entity.Clients, err error) {
	result = entity.Clients{}
	stmt := Db()

	search = strings.TrimSpace(search)

	if search == "all" {
		// Don't filter.
	} else if rnd.IsUID(search, entity.ClientUID) {
		stmt = stmt.Where("client_uid = ?", search)
	} else if rnd.IsUID(search, entity.UserUID) {
		stmt = stmt.Where("user_uid = ?", search)
	} else if search != "" {
		stmt = stmt.Where("client_name LIKE ? OR user_name LIKE ?", search+"%", search+"%")
	}

	if sortOrder == "" {
		sortOrder = "user_name, client_name, created_at"
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
