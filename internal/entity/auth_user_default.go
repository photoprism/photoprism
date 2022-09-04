package entity

import (
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const AdminUserName = "admin"
const AdminDisplayName = "Admin"
const GuestDisplayName = "Guest"

// Admin is the default admin user.
var Admin = User{
	ID:          1,
	UserSlug:    "admin",
	Username:    AdminUserName,
	UserRole:    acl.RoleAdmin.String(),
	DisplayName: AdminDisplayName,
	SuperAdmin:  true,
	CanLogin:    true,
	CanInvite:   true,
	InviteToken: rnd.GenerateToken(8),
}

// UnknownUser is an anonymous, public user without own account.
var UnknownUser = User{
	ID:          -1,
	UserSlug:    "",
	UserUID:     "u000000000000001",
	UserRole:    "",
	Username:    "",
	DisplayName: "",
	SuperAdmin:  false,
	CanLogin:    false,
	CanInvite:   false,
	InviteToken: "",
}

// Guest is a user without own account e.g. for link sharing.
var Guest = User{
	ID:          -2,
	UserSlug:    "guest",
	UserUID:     "u000000000000002",
	UserRole:    acl.RoleGuest.String(),
	Username:    "guest",
	DisplayName: GuestDisplayName,
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
