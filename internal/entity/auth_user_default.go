package entity

import (
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Role defaults.
const (
	AdminUserName    = "admin"
	AdminDisplayName = "Admin"
	GuestDisplayName = "Visitor"
)

// Admin is the default admin user.
var Admin = User{
	ID:          1,
	UserName:    AdminUserName,
	UserRole:    acl.RoleAdmin.String(),
	DisplayName: AdminDisplayName,
	SuperAdmin:  true,
	CanLogin:    true,
	CanSync:     true,
	CanInvite:   true,
	InviteToken: rnd.GenerateToken(8),
}

// UnknownUser is an anonymous, public user without own account.
var UnknownUser = User{
	ID:          -1,
	UserUID:     "u000000000000001",
	UserRole:    acl.RoleUnauthorized.String(),
	UserName:    "",
	DisplayName: "",
	SuperAdmin:  false,
	CanLogin:    false,
	CanSync:     false,
	CanInvite:   false,
	InviteToken: "",
}

// Visitor is a user without own account e.g. for link sharing.
var Visitor = User{
	ID:          -2,
	UserUID:     "u000000000000002",
	UserRole:    acl.RoleVisitor.String(),
	UserName:    "",
	DisplayName: GuestDisplayName,
	SuperAdmin:  false,
	CanLogin:    false,
	CanSync:     false,
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

	if user := FirstOrCreateUser(&Visitor); user != nil {
		Visitor = *user
	}
}
