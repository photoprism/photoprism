package entity

import (
	"github.com/photoprism/photoprism/pkg/clean"
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
	if ctx.IsSet("name") {
		m.DisplayName = clean.Name(frm.DisplayName)
	}

	// User role.
	if ctx.IsSet("role") {
		m.SetRole(frm.Role())
	}

	// Super-admin status.
	if ctx.IsSet("superadmin") {
		m.SuperAdmin = frm.SuperAdmin
	}

	// Disable login (Web UI)?
	if ctx.IsSet("no-login") {
		m.CanLogin = frm.CanLogin
	}

	// Allow the use of WebDAV?
	if ctx.IsSet("webdav") {
		m.WebDAV = frm.WebDAV
	}

	// Set custom attributes?
	if ctx.IsSet("attr") {
		m.UserAttr = frm.Attr()
	}

	// Originals base folder.
	if ctx.IsSet("base-path") {
		m.SetBasePath(frm.BasePath)
	}

	// Sub-folder for uploads.
	if ctx.IsSet("upload-path") {
		m.SetUploadPath(frm.UploadPath)
	}

	return m.Validate()
}

// RestoreFromCli restored the account from a CLI context.
func (m *User) RestoreFromCli(ctx *cli.Context, newPassword string) (err error) {
	m.DeletedAt = nil

	// Set values.
	if err = m.SetValuesFromCli(ctx); err != nil {
		return err
	}

	// Save values.
	if err = m.Save(); err != nil {
		return err
	} else if newPassword == "" {
		return nil
	} else if err = m.SetPassword(newPassword); err != nil {
		return err
	}

	return nil
}
