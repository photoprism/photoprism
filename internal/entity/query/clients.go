package query

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Clients finds clients and returns them. When deleted is true it reports the retired
// records instead of the current ones, so an operator can see what a deletion left behind.
func Clients(limit, offset int, sortOrder, search string, deleted bool) (result entity.Clients, err error) {
	result = entity.Clients{}
	stmt := Db()

	if deleted {
		stmt = UnscopedDb().Where("deleted_at IS NOT NULL")
	}

	search = strings.TrimSpace(search)

	switch {
	case search == "all":
		// Don't filter.
	case rnd.IsUID(search, entity.ClientUID):
		stmt = stmt.Where("client_uid = ?", search)
	case rnd.IsUID(search, entity.UserUID):
		stmt = stmt.Where("user_uid = ?", search)
	case search != "":
		switch entity.DbDialect() {
		case dsn.DialectPostgreSQL:
			stmt = stmt.Where("lower(client_name) LIKE ? OR lower(user_name) LIKE ?", strings.ToLower(search+"%"), strings.ToLower(search+"%"))
		default:
			stmt = stmt.Where("client_name LIKE ? OR user_name LIKE ?", search+"%", search+"%")
		}
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
