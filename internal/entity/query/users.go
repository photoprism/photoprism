package query

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/convert"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// RegisteredUsers finds all registered users.
func RegisteredUsers() (result entity.Users) {
	if err := Db().Where("id > 0").Find(&result).Error; err != nil {
		log.Errorf("users: %s (find)", err)
	}

	return result
}

// CountUsers returns the number of users based on the specified filter options.
func CountUsers(registered, active bool, includeRoles, excludeRoles []string) (count int) {
	if !Db().Migrator().HasTable("auth_users") {
		return 0
	}

	countData := int64(0)
	stmt := Db().Model(&entity.Users{})

	if registered {
		stmt = stmt.Where("id > 0")
	}

	if active {
		stmt = stmt.Where("(can_login = TRUE OR webdav = TRUE) AND user_name <> ''")
	}

	if len(includeRoles) > 0 {
		stmt = stmt.Where("user_role IN (?)", includeRoles)
	} else if len(excludeRoles) > 0 {
		stmt = stmt.Where("user_role NOT IN (?)", excludeRoles)
	}

	if err := stmt.Count(&countData).Error; err != nil {
		log.Errorf("users: %s (count)", err)
	}

	return convert.SafeInt64toint(countData)
}

// Users finds user accounts based on the specified parameters.
func Users(limit, offset int, sortOrder, search string, deleted bool) (result entity.Users, err error) {
	result = entity.Users{}
	stmt := UnscopedDb()

	search = strings.TrimSpace(search)

	// Filter user accounts to be returned.
	if search == "all" {
		// Don't filter.
	} else if id := txt.Int(search); id != 0 {
		stmt = stmt.Where("id = ?", id)
	} else if rnd.IsUID(search, entity.UserUID) {
		stmt = stmt.Where("user_uid = ?", search)
	} else if search != "" {
		switch entity.DbDialect() {
		case dsn.DialectPostgreSQL:
			lowerSearch := strings.ToLower(search + "%")
			stmt = stmt.Where("lower(user_name) LIKE ? OR lower(user_email) LIKE ? OR lower(display_name) LIKE ?", lowerSearch, lowerSearch, lowerSearch)
		default:
			stmt = stmt.Where("user_name LIKE ? OR user_email LIKE ? OR display_name LIKE ?", search+"%", search+"%", search+"%")
		}

	} else {
		stmt = stmt.Where("id > 0")
	}

	// Hide deleted user accounts?
	if !deleted {
		stmt = stmt.Where("deleted_at IS NULL")
	}

	// Set result sort order.
	if sortOrder == "" {
		sortOrder = "id"
	}

	// Limit number of results.
	if limit > 0 {
		stmt = stmt.Limit(limit)

		if offset > 0 {
			stmt = stmt.Offset(offset)
		}
	}

	err = stmt.Order(sortOrder).Find(&result).Error

	return result, err
}
