package entity

import (
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const DefaultAdminUserName = "admin"
const DefaultAdminFullName = "Admin"

const DisplayNameUnknown = "Public"
const DisplayNameGuest = "Guest"

// Admin is the default admin user.
var Admin = User{
	ID:          1,
	UserSlug:    "admin",
	Username:    DefaultAdminUserName,
	UserRole:    acl.RoleAdmin.String(),
	DisplayName: DefaultAdminFullName,
	SuperAdmin:  true,
	CanLogin:    true,
	CanInvite:   true,
	InviteToken: rnd.GenerateToken(8),
}

// UnknownUser is an anonymous, public user without own account.
var UnknownUser = User{
	ID:          -1,
	UserSlug:    "1",
	UserUID:     "u000000000000001",
	UserRole:    "",
	Username:    "",
	DisplayName: DisplayNameUnknown,
	SuperAdmin:  false,
	CanLogin:    false,
	CanInvite:   false,
	InviteToken: "",
}

// Guest is a user without own account e.g. for link sharing.
var Guest = User{
	ID:          -2,
	UserSlug:    "2",
	UserUID:     "u000000000000002",
	UserRole:    acl.RoleGuest.String(),
	Username:    "",
	DisplayName: DisplayNameGuest,
	SuperAdmin:  false,
	CanLogin:    false,
	CanInvite:   false,
	InviteToken: "",
}

// CreateDefaultUsers initializes the database with default user accounts.
func CreateDefaultUsers() {
	if user := FirstOrCreateUser(&Admin); user != nil {
		Admin = *user
	}

	if user := FirstOrCreateUser(&UnknownUser); user != nil {
		UnknownUser = *user
	}

	if user := FirstOrCreateUser(&Guest); user != nil {
		Guest = *user
	}
}
