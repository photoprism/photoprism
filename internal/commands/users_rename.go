package commands

import (
	"slices"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// UserUsernameUsage is the usage hint of the users mod --username flag.
const UserUsernameUsage = "changes the login `USERNAME`, the account keeps its folders and sessions"

// UserUsernameFlag returns the --username flag of the users mod commands.
func UserUsernameFlag() *cli.StringFlag {
	return &cli.StringFlag{Name: "username", Usage: UserUsernameUsage}
}

// PrepareUserRename sets the name given with --username, or else the normalized form of the stored
// name, before the other flags are applied, and returns the name it replaces. A refused name changes
// nothing, so it is checked before SetValuesFromCli can revoke sessions.
func PrepareUserRename(ctx *cli.Context, m *entity.User) (oldName string, err error) {
	oldName = m.UserName

	if ctx.IsSet("username") {
		err = m.SetNewUsername(ctx.String("username"))
	} else {
		m.SetNormalizedUsername()
	}

	return oldName, err
}

// SaveUserRename saves the account and the copies of a changed username in one transaction, and
// reports the rename. A forced rename is audited as a warning.
func SaveUserRename(ctx *cli.Context, conf *config.Config, m *entity.User, oldName string, forced bool) error {
	if err := m.SaveRename(oldName); err != nil {
		return err
	}

	if m.UserName != oldName {
		log.Infof("user %s has been renamed from %s", clean.LogQuote(m.UserName), clean.LogQuote(oldName))

		msg := []string{"cli", "user %s", "change username from %s to %s", status.Succeeded}

		if forced {
			event.AuditWarn(msg, m.RefID, clean.LogQuote(oldName), clean.LogQuote(m.UserName))
		} else {
			event.AuditInfo(msg, m.RefID, clean.LogQuote(oldName), clean.LogQuote(m.UserName))
		}

		WarnOidcEmailUsername(conf, m)
		WarnRenamedFolder(m, oldName)
	}

	if !ctx.IsSet("username") || UserFlagsSet(ctx, "username", "force") {
		log.Infof("user %s has been updated", m.String())
	}

	return nil
}

// WarnOidcEmailUsername logs a warning and returns true if the username of an OIDC account is an
// email address other than the account's own while OIDC usernames are taken from email addresses.
func WarnOidcEmailUsername(conf *config.Config, m *entity.User) bool {
	if conf == nil || m == nil || conf.OIDCUsername() != authn.OidcClaimEmail || !m.HasProvider(authn.ProviderOIDC) {
		return false
	} else if !strings.Contains(m.UserName, "@") || strings.EqualFold(m.UserName, m.UserEmail) {
		return false
	}

	log.Warnf("user %s: username does not match the email address, which OIDC uses as username", clean.LogQuote(m.UserName))

	return true
}

// WarnRenamedFolder logs a warning and returns true if the account keeps using the default folder of
// its former username while that name was freed or its default folder changed.
func WarnRenamedFolder(m *entity.User, oldName string) bool {
	if m == nil || oldName == "" {
		return false
	}

	former := &entity.User{UserUID: m.UserUID, UserName: oldName}
	folder := former.DefaultBasePath()

	if !former.ValidHandle() || clean.Username(oldName) == m.UserName && folder == m.DefaultBasePath() {
		return false
	}

	effective := *m
	basePath, uploadPath := effective.GetBasePath(), effective.GetUploadPath()

	if basePath != folder && uploadPath != folder && !strings.HasPrefix(uploadPath, folder+"/") {
		return false
	}

	log.Warnf("user %s keeps the folder %s, assign another base path to new accounts with the handle %s", clean.LogQuote(m.UserName), clean.Log(folder), clean.LogQuote(former.Handle()))

	return true
}

// UserFlagsSet reports whether a flag of the command other than the ignored ones was set.
func UserFlagsSet(ctx *cli.Context, ignore ...string) bool {
	for _, name := range ctx.LocalFlagNames() {
		if !slices.Contains(ignore, name) {
			return true
		}
	}

	return false
}
