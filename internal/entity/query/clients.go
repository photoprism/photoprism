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

	switch {
	case search == "all":
		// Don't filter.
	case rnd.IsUID(search, entity.ClientUID):
		stmt = stmt.Where("client_uid = ?", search)
	case rnd.IsUID(search, entity.UserUID):
		stmt = stmt.Where("user_uid = ?", search)
	case search != "":
		stmt = stmt.Where("client_name LIKE ? OR user_name LIKE ?", search+"%", search+"%")
	}

	if sortOrder == "" {
		sortOrder = "created_at DESC, client_uid DESC"
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
