package entity

import (
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/form"
)

// SetValuesFromCli updates the entity values from a CLI context and validates them.
func (m *User) SetValuesFromCli(ctx *cli.Context) error {
	frm := form.NewUserFromCli(ctx)

	// Email address.
	if ctx.IsSet("email") {
		m.UserEmail = frm.Email()
	}

	// Display name.
	if ctx.IsSet("displayname") {
		m.DisplayName = frm.DisplayName
	}

	// User role.
	if ctx.IsSet("role") {
		m.UserRole = frm.Role()
	}

	// Custom attributes.
	if ctx.IsSet("attr") {
		m.UserAttr = frm.Attr()
	}

	// Super-admin status.
	if ctx.IsSet("superadmin") {
		m.SuperAdmin = frm.SuperAdmin
	}

	// Disable Web UI?
	if ctx.IsSet("disable-login") {
		m.CanLogin = frm.CanLogin
	}

	// Can use WebDAV.
	if ctx.IsSet("can-sync") {
		m.CanSync = frm.CanSync
	}

	return m.Validate()
}
