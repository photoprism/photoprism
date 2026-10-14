package entity

import (
	"strings"

	"github.com/photoprism/photoprism/internal/entity/legacy"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// FindLegacyUser returns the matching legacy user or nil if it was not found.
func FindLegacyUser(find User) *legacy.User {
	m := &legacy.User{}

	// Build query.
	stmt := Db()
	if find.ID != 0 {
		stmt = stmt.Where("id = ?", find.ID)
	} else if find.UserUID != "" {
		stmt = stmt.Where("user_uid = ?", find.UserUID)
	} else if find.UserName != "" {
		stmt = stmt.Where("user_name = ?", find.UserName)
	} else if find.UserEmail != "" {
		stmt = stmt.Where("primary_email = ?", find.UserEmail)
	} else {
		return nil
	}

	if found, err := DbHasTable(Db(), m.TableName()); err != nil || !found {
		return nil
	}

	// Find matching record.
	if err := stmt.First(m).Error; err != nil {
		return nil
	}

	return m
}

// FindLegacyUsers finds registered legacy users.
func FindLegacyUsers(search string) legacy.Users {
	result := legacy.Users{}

	if found, err := DbHasTable(Db(), legacy.User{}.TableName()); err != nil || !found {
		return result
	}

	stmt := Db()

	search = strings.TrimSpace(search)

	if search == "all" {
		// Don't filter.
	} else if id := txt.Int(search); id != 0 {
		stmt = stmt.Where("id = ?", id)
	} else if rnd.IsUID(search, UserUID) {
		stmt = stmt.Where("user_uid = ?", search)
	} else if search != "" {
		like := clean.SqlLike(search) + "%"
		stmt = stmt.Where(clean.SqlLikeAny(DbDialect(), false, "user_name", "primary_email", "full_name"), like, like, like)
	} else {
		stmt = stmt.Where("id > 0")
	}

	stmt.Order("id").Find(&result)

	return result
}
