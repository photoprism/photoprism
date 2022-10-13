package entity

import "github.com/photoprism/photoprism/internal/entity/legacy"

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

	// Find matching record.
	if err := stmt.First(m).Error; err != nil {
		return nil
	}

	return m
}

// FindLegacyUsers finds registered legacy users.
func FindLegacyUsers() legacy.Users {
	result := make(legacy.Users, 0, 1)

	Db().Where("id > 0").Find(&result)

	return result
}
