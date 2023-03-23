package entity

import (
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
)

// Role defaults.
const (
	AdminUserName      = "admin"
	AdminDisplayName   = "Admin"
	VisitorDisplayName = "Visitor"
	UnknownDisplayName = "Unknown"
)

// Admin is the default admin user.
var Admin = User{
	ID:            1,
	UserName:      AdminUserName,
	AuthProvider:  authn.ProviderLocal.String(),
	UserRole:      acl.RoleAdmin.String(),
	DisplayName:   AdminDisplayName,
	SuperAdmin:    true,
	CanLogin:      true,
	WebDAV:        true,
	CanInvite:     true,
	InviteToken:   GenerateToken(),
	PreviewToken:  GenerateToken(),
	DownloadToken: GenerateToken(),
}

// UnknownUser is an anonymous, public user without own account.
var UnknownUser = User{
	ID:            -1,
	UserUID:       "u000000000000001",
	UserName:      "",
	AuthProvider:  authn.ProviderNone.String(),
	UserRole:      acl.RoleUnknown.String(),
	CanLogin:      false,
	WebDAV:        false,
	CanInvite:     false,
	DisplayName:   UnknownDisplayName,
	InviteToken:   "",
	PreviewToken:  "",
	DownloadToken: "",
}

// Visitor is a user without own account e.g. for link sharing.
var Visitor = User{
	ID:            -2,
	UserUID:       "u000000000000002",
	UserName:      "",
	AuthProvider:  authn.ProviderLink.String(),
	UserRole:      acl.RoleVisitor.String(),
	DisplayName:   VisitorDisplayName,
	CanLogin:      false,
	WebDAV:        false,
	CanInvite:     false,
	InviteToken:   "",
	PreviewToken:  "",
	DownloadToken: "",
}

// CreateDefaultUsers initializes the database with default user accounts.
func CreateDefaultUsers() {
	if admin := FindUser(Admin); admin != nil {
		Admin = *admin
	} else {
		// Set legacy values.
		if leg := FindLegacyUser(Admin); leg != nil {
			Admin.UserUID = leg.UserUID
			if leg.UserName != "" {
				Admin.UserName = leg.UserName
			}
			if leg.PrimaryEmail != "" {
				Admin.UserEmail = leg.PrimaryEmail
			}
			if leg.FullName != "" {
				Admin.DisplayName = leg.FullName
			}
			if leg.LoginAt != nil {
				Admin.LoginAt = leg.LoginAt
			}
			log.Infof("users: migrating %s account", Admin.UserName)
		}

		// Set default values.
		Admin.SuperAdmin = true
		Admin.CanLogin = true
		Admin.WebDAV = true

		// Username is required.
		if Admin.UserName == "" {
			Admin.UserName = "admin"
		}

		// Add initial admin account.
		if err := Admin.Create(); err != nil {
			event.AuditErr([]string{"user", "failed to create", "%s"}, err)
		}
	}

	if user := FirstOrCreateUser(&UnknownUser); user != nil {
		UnknownUser = *user
	}

	if user := FirstOrCreateUser(&Visitor); user != nil {
		Visitor = *user
	}
}
