package entity

import (
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
)

// SetValuesFromCli updates the entity values from a CLI context and validates them.
func (m *User) SetValuesFromCli(ctx *cli.Context) error {
	frm := form.NewUserFromCli(ctx)

	// see https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html#renew-the-session-id-after-any-privilege-level-change
	privilegeLevelChange := false

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
		privilegeLevelChange = true
	}

	// Super-admin status.
	if ctx.IsSet("superadmin") {
		m.SuperAdmin = frm.SuperAdmin
		privilegeLevelChange = true
	}

	// Disable login (Web UI)?
	if ctx.IsSet("no-login") {
		m.CanLogin = frm.CanLogin
		privilegeLevelChange = true
	}

	// Allow the use of WebDAV?
	if ctx.IsSet("webdav") {
		m.WebDAV = frm.WebDAV
		privilegeLevelChange = true
	}

	// Set custom attributes?
	if ctx.IsSet("attr") {
		m.UserAttr = frm.Attr()
		privilegeLevelChange = true
	}

	// Originals base folder.
	if ctx.IsSet("base-path") {
		m.SetBasePath(frm.BasePath)
		privilegeLevelChange = true
	}

	// Sub-folder for uploads.
	if ctx.IsSet("upload-path") {
		m.SetUploadPath(frm.UploadPath)
		privilegeLevelChange = true
	}

	// Disable two-factor authentication.
	if ctx.IsSet("disable-2fa") && m.Method().Is(authn.Method2FA) {
		m.SetMethod(authn.MethodDefault)
		privilegeLevelChange = true
	}

	// Validate properties.
	if err := m.Validate(); err != nil {
		// Invalid.
		return err
	} else if privilegeLevelChange {
		// Delete sessions after privilege level change.
		m.DeleteSessions(nil)
	}

	return nil
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
