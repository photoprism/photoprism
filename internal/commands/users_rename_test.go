package commands

import (
	"errors"
	"strings"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
)

// newRenameTestUser stores an account with the given name and a login session, and removes both after the test.
func newRenameTestUser(t *testing.T, name string) (*entity.User, *entity.Session) {
	t.Helper()

	m := &entity.User{UserName: name, UserRole: "admin", CanLogin: true}
	require.NoError(t, m.Create())

	sess := entity.NewSession(3600, 0)
	sess.SetUser(m)
	sess.SetProvider(authn.ProviderLocal)
	require.NoError(t, sess.Create())

	t.Cleanup(func() {
		reopenConnection()
		assert.NoError(t, entity.UnscopedDb().Delete(&entity.Session{}, "user_uid = ?", m.UserUID).Error)
		removeTestUser(t, m.UserUID)
	})

	return m, sess
}

// renameSessionUserName returns the username stored with the session, or an empty string if it was deleted.
func renameSessionUserName(t *testing.T, id string) string {
	t.Helper()

	var found entity.Session

	if err := entity.UnscopedDb().Where("id = ?", id).First(&found).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return ""
	} else {
		require.NoError(t, err)
	}

	return found.UserName
}

// captureCommandLog replaces the command logger for the duration of a test.
func captureCommandLog(t *testing.T) *test.Hook {
	t.Helper()

	orig := log
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	log = logger

	t.Cleanup(func() { log = orig })

	return hook
}

// logMessages returns the formatted messages of the captured entries at the given level.
func logMessages(hook *test.Hook, level logrus.Level) (messages []string) {
	for _, entry := range hook.AllEntries() {
		if entry.Level == level {
			messages = append(messages, entry.Message)
		}
	}

	return messages
}

// renameAuditEntry returns the audit entry that records a username change, or nil if there is none.
func renameAuditEntry(hook *test.Hook) *logrus.Entry {
	for _, entry := range hook.AllEntries() {
		if strings.Contains(entry.Message, "change username from") {
			return entry
		}
	}

	return nil
}

func TestUsersModCommand_Username(t *testing.T) {
	requireTestDb(t)

	t.Run("Success", func(t *testing.T) {
		m, sess := newRenameTestUser(t, "modrename-old")
		audit := captureAuditLog(t)
		logs := captureCommandLog(t)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--username=ModRename-New", "modrename-old"})
		require.NoError(t, err)

		renamed := entity.FindUserByUID(m.UserUID)
		require.NotNil(t, renamed)
		assert.Equal(t, "modrename-new", renamed.UserName)
		assert.Nil(t, entity.FindUserByName("modrename-old"))
		assert.Equal(t, "modrename-new", renameSessionUserName(t, sess.ID))

		entry := renameAuditEntry(audit)
		require.NotNil(t, entry)
		assert.Equal(t, logrus.InfoLevel, entry.Level)
		assert.Contains(t, entry.Message, "'modrename-old' to 'modrename-new'")
		assert.Contains(t, logMessages(logs, logrus.InfoLevel), "user 'modrename-new' has been renamed from 'modrename-old'")
		assert.NotContains(t, strings.Join(logMessages(logs, logrus.InfoLevel), "\n"), "has been updated")
	})
	t.Run("WithOtherFlagsByUID", func(t *testing.T) {
		m, sess := newRenameTestUser(t, "modrename-uid")
		logs := captureCommandLog(t)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--username=modrename-uid-new", "--email=modrename@example.com", "--name=Renamed By UID", m.UserUID})
		require.NoError(t, err)

		renamed := entity.FindUserByUID(m.UserUID)
		require.NotNil(t, renamed)
		assert.Equal(t, "modrename-uid-new", renamed.UserName)
		assert.Equal(t, "modrename@example.com", renamed.UserEmail)
		assert.Equal(t, "Renamed By UID", renamed.DisplayName)
		assert.Equal(t, "modrename-uid-new", renameSessionUserName(t, sess.ID))
		assert.Contains(t, strings.Join(logMessages(logs, logrus.InfoLevel), "\n"), "has been updated")
	})
	t.Run("Restore", func(t *testing.T) {
		m, _ := newRenameTestUser(t, "modrename-deleted")
		require.NoError(t, m.Delete())

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--restore", "--username=modrename-restored", "modrename-deleted"})
		require.NoError(t, err)

		renamed := entity.FindUserByUID(m.UserUID)
		require.NotNil(t, renamed)
		assert.False(t, renamed.IsDeleted())
		assert.Equal(t, "modrename-restored", renamed.UserName)
	})
	t.Run("SameName", func(t *testing.T) {
		m, sess := newRenameTestUser(t, "modrename-same")
		audit := captureAuditLog(t)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--username=ModRename-Same", "modrename-same"})
		require.NoError(t, err)

		assert.Equal(t, "modrename-same", entity.FindUserByUID(m.UserUID).UserName)
		assert.Equal(t, "modrename-same", renameSessionUserName(t, sess.ID))
		assert.Nil(t, renameAuditEntry(audit))
	})
	t.Run("RefusedBeforeRevoke", func(t *testing.T) {
		m, sess := newRenameTestUser(t, "modrename-refused")
		audit := captureAuditLog(t)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--role=guest", "--username=alice", "modrename-refused"})
		require.EqualError(t, err, "user 'alice' already exists")

		stored := entity.FindUserByUID(m.UserUID)
		assert.Equal(t, "modrename-refused", stored.UserName)
		assert.Equal(t, "admin", stored.UserRole)
		assert.Equal(t, "modrename-refused", renameSessionUserName(t, sess.ID))
		assert.Nil(t, renameAuditEntry(audit))
	})
	t.Run("Invalid", func(t *testing.T) {
		m, _ := newRenameTestUser(t, "modrename-invalid")

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--username=visitor", "modrename-invalid"})
		require.EqualError(t, err, "username is reserved")
		assert.Equal(t, "modrename-invalid", entity.FindUserByUID(m.UserUID).UserName)
	})
	t.Run("FailedRename", func(t *testing.T) {
		m, sess := newRenameTestUser(t, "modrename-failed")
		entity.UserRenameHandlers["zz_rename_test"] = func(*gorm.DB, *entity.User, string, string) error {
			return errors.New("test error")
		}
		t.Cleanup(func() { delete(entity.UserRenameHandlers, "zz_rename_test") })

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Not Applied", "--username=modrename-failed-new", "modrename-failed"})
		require.EqualError(t, err, "failed to update zz_rename_test (test error)")

		stored := entity.FindUserByUID(m.UserUID)
		assert.Equal(t, "modrename-failed", stored.UserName)
		assert.NotEqual(t, "Not Applied", stored.DisplayName)
		assert.Equal(t, "modrename-failed", renameSessionUserName(t, sess.ID))
	})
	t.Run("TrailingFlag", func(t *testing.T) {
		m, _ := newRenameTestUser(t, "modrename-trailing")

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "modrename-trailing", "--username", "modrename-trailing-new"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must appear before positional arguments")
		assert.Equal(t, "modrename-trailing", entity.FindUserByUID(m.UserUID).UserName)
	})
	t.Run("LegacyName", func(t *testing.T) {
		m, sess := newRenameTestUser(t, "modrename-legacy")
		require.NoError(t, entity.UnscopedDb().Model(&entity.User{}).Where("id = ?", m.ID).UpdateColumn("user_name", "ModRename-Legacy").Error)
		require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("id = ?", sess.ID).UpdateColumn("user_name", "ModRename-Legacy").Error)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Legacy Name", m.UserUID})
		require.NoError(t, err)

		stored := entity.FindUserByUID(m.UserUID)
		assert.Equal(t, "modrename-legacy", stored.UserName)
		assert.Equal(t, "Legacy Name", stored.DisplayName)
		assert.Equal(t, "modrename-legacy", renameSessionUserName(t, sess.ID))
	})
	t.Run("LegacyNamePaths", func(t *testing.T) {
		m, _ := newRenameTestUser(t, "modrename-paths")
		require.NoError(t, entity.UnscopedDb().Model(&entity.User{}).Where("id = ?", m.ID).
			Updates(entity.Values{"user_name": "modrename+paths", "upload_path": "~"}).Error)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Legacy Paths", m.UserUID})
		require.NoError(t, err)

		stored := entity.FindUserByUID(m.UserUID)
		assert.Equal(t, "modrenamepaths", stored.UserName)
		assert.Equal(t, "users/modrename.paths", stored.UploadPath)
	})
	t.Run("LegacyNameHeldByActive", func(t *testing.T) {
		newRenameTestUser(t, "modrename-active")
		m, _ := newRenameTestUser(t, "modrename-active-legacy")
		require.NoError(t, entity.UnscopedDb().Model(&entity.User{}).Where("id = ?", m.ID).UpdateColumn("user_name", "ModRename-Active").Error)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Active Legacy", m.UserUID})
		require.EqualError(t, err, "user 'modrename-active' already exists")
		assert.Equal(t, "ModRename-Active", entity.FindUserByUID(m.UserUID).UserName)
	})
	t.Run("LegacyNameHeldByDeleted", func(t *testing.T) {
		holder, _ := newRenameTestUser(t, "modrename-held")
		require.NoError(t, holder.Delete())
		m, sess := newRenameTestUser(t, "modrename-held-legacy")
		require.NoError(t, entity.UnscopedDb().Model(&entity.User{}).Where("id = ?", m.ID).UpdateColumn("user_name", "ModRename-Held").Error)
		require.NoError(t, entity.UnscopedDb().Model(&entity.Session{}).Where("id = ?", sess.ID).UpdateColumn("user_name", "ModRename-Held").Error)
		audit := captureAuditLog(t)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Held Legacy", m.UserUID})
		require.NoError(t, err)

		stored := entity.FindUserByUID(m.UserUID)
		assert.Equal(t, "ModRename-Held", stored.UserName)
		assert.Equal(t, "Held Legacy", stored.DisplayName)
		assert.Equal(t, "ModRename-Held", renameSessionUserName(t, sess.ID))
		assert.Nil(t, renameAuditEntry(audit))
	})
	t.Run("Unchanged", func(t *testing.T) {
		m, sess := newRenameTestUser(t, "modrename-plain")
		audit := captureAuditLog(t)
		logs := captureCommandLog(t)

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Plain Change", "modrename-plain"})
		require.NoError(t, err)

		stored := entity.FindUserByUID(m.UserUID)
		assert.Equal(t, "modrename-plain", stored.UserName)
		assert.Equal(t, "Plain Change", stored.DisplayName)
		assert.Equal(t, "modrename-plain", renameSessionUserName(t, sess.ID))
		assert.Nil(t, renameAuditEntry(audit))
		assert.Contains(t, logMessages(logs, logrus.InfoLevel), "user 'modrename-plain' has been updated")
	})
}

func TestPrepareUserRename(t *testing.T) {
	requireTestDb(t)

	t.Run("Success", func(t *testing.T) {
		m, _ := newRenameTestUser(t, "preprename-old")

		oldName, err := PrepareUserRename(newCommandContext(t, UsersModCommand, "--username=preprename-new"), m)
		require.NoError(t, err)
		assert.Equal(t, "preprename-old", oldName)
		assert.Equal(t, "preprename-new", m.UserName)
		assert.Equal(t, "preprename-old", entity.FindUserByUID(m.UserUID).UserName)
	})
	t.Run("Refused", func(t *testing.T) {
		m, _ := newRenameTestUser(t, "preprename-refused")

		oldName, err := PrepareUserRename(newCommandContext(t, UsersModCommand, "--username=alice"), m)
		require.EqualError(t, err, "user 'alice' already exists")
		assert.Equal(t, "preprename-refused", oldName)
		assert.Equal(t, "preprename-refused", m.UserName)
	})
	t.Run("Normalized", func(t *testing.T) {
		m, _ := newRenameTestUser(t, "preprename-legacy")
		m.UserName = "PrepRename-Legacy"

		oldName, err := PrepareUserRename(newCommandContext(t, UsersModCommand, "--name=Legacy"), m)
		require.NoError(t, err)
		assert.Equal(t, "PrepRename-Legacy", oldName)
		assert.Equal(t, "preprename-legacy", m.UserName)
	})
	t.Run("NoFlags", func(t *testing.T) {
		m, _ := newRenameTestUser(t, "preprename-none")

		oldName, err := PrepareUserRename(newCommandContext(t, UsersModCommand), m)
		require.NoError(t, err)
		assert.Equal(t, "preprename-none", oldName)
		assert.Equal(t, "preprename-none", m.UserName)
	})
}

func TestSaveUserRename(t *testing.T) {
	requireTestDb(t)

	t.Run("Forced", func(t *testing.T) {
		m, sess := newRenameTestUser(t, "saverename-forced")
		audit := captureAuditLog(t)
		ctx := newCommandContext(t, UsersModCommand, "--username=saverename-forced-new")

		oldName, err := PrepareUserRename(ctx, m)
		require.NoError(t, err)
		require.NoError(t, SaveUserRename(ctx, get.Config(), m, oldName, true))

		entry := renameAuditEntry(audit)
		require.NotNil(t, entry)
		assert.Equal(t, logrus.WarnLevel, entry.Level)
		assert.Equal(t, "saverename-forced-new", entity.FindUserByUID(m.UserUID).UserName)
		assert.Equal(t, "saverename-forced-new", renameSessionUserName(t, sess.ID))
	})
	t.Run("Warnings", func(t *testing.T) {
		conf := get.Config()
		orig := conf.Options().OIDCUsername
		conf.Options().OIDCUsername = authn.OidcClaimEmail
		t.Cleanup(func() { conf.Options().OIDCUsername = orig })

		m, _ := newRenameTestUser(t, "saverename-warn@old.example")
		m.UploadPath = "~"
		m.AuthProvider = authn.ProviderOIDC.String()
		m.UserEmail = "saverename-warn@old.example"
		logs := captureCommandLog(t)
		ctx := newCommandContext(t, UsersModCommand, "--username=saverename-other@new.example")

		oldName, err := PrepareUserRename(ctx, m)
		require.NoError(t, err)
		require.NoError(t, SaveUserRename(ctx, conf, m, oldName, false))
		assert.Len(t, logMessages(logs, logrus.WarnLevel), 2)
	})
	t.Run("Unchanged", func(t *testing.T) {
		m, _ := newRenameTestUser(t, "saverename-unchanged")
		audit := captureAuditLog(t)
		ctx := newCommandContext(t, UsersModCommand, "--name=Unchanged")

		m.DisplayName = "Unchanged"
		require.NoError(t, SaveUserRename(ctx, get.Config(), m, m.UserName, false))
		assert.Nil(t, renameAuditEntry(audit))
		assert.Equal(t, "Unchanged", entity.FindUserByUID(m.UserUID).DisplayName)
	})
	t.Run("Error", func(t *testing.T) {
		m, _ := newRenameTestUser(t, "saverename-error")
		audit := captureAuditLog(t)
		ctx := newCommandContext(t, UsersModCommand, "--username=saverename-error-new")
		entity.UserRenameHandlers["zz_rename_test"] = func(*gorm.DB, *entity.User, string, string) error {
			return errors.New("test error")
		}
		t.Cleanup(func() { delete(entity.UserRenameHandlers, "zz_rename_test") })

		oldName, err := PrepareUserRename(ctx, m)
		require.NoError(t, err)
		require.Error(t, SaveUserRename(ctx, get.Config(), m, oldName, false))
		assert.Nil(t, renameAuditEntry(audit))
	})
}

func TestWarnOidcEmailUsername(t *testing.T) {
	conf := requireTestDb(t)
	orig := conf.Options().OIDCUsername
	t.Cleanup(func() { conf.Options().OIDCUsername = orig })

	oidcUser := func(name, email string) *entity.User {
		return &entity.User{UserName: name, UserEmail: email, AuthProvider: authn.ProviderOIDC.String()}
	}

	conf.Options().OIDCUsername = authn.OidcClaimEmail

	t.Run("Differs", func(t *testing.T) {
		logs := captureCommandLog(t)
		assert.True(t, WarnOidcEmailUsername(conf, oidcUser("jane@new.example", "jane@old.example")))
		assert.Len(t, logMessages(logs, logrus.WarnLevel), 1)
	})
	t.Run("Matches", func(t *testing.T) {
		assert.False(t, WarnOidcEmailUsername(conf, oidcUser("jane@new.example", "Jane@New.example")))
	})
	t.Run("NoEmail", func(t *testing.T) {
		assert.False(t, WarnOidcEmailUsername(conf, oidcUser("jane", "jane@new.example")))
	})
	t.Run("LocalAccount", func(t *testing.T) {
		m := oidcUser("jane@new.example", "jane@old.example")
		m.AuthProvider = authn.ProviderLocal.String()
		assert.False(t, WarnOidcEmailUsername(conf, m))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, WarnOidcEmailUsername(nil, oidcUser("jane@new.example", "jane@old.example")))
		assert.False(t, WarnOidcEmailUsername(conf, nil))
	})
	t.Run("PreferredUsername", func(t *testing.T) {
		conf.Options().OIDCUsername = authn.OidcClaimPreferredUsername
		t.Cleanup(func() { conf.Options().OIDCUsername = authn.OidcClaimEmail })
		assert.False(t, WarnOidcEmailUsername(conf, oidcUser("jane@new.example", "jane@old.example")))
	})
}

func TestUserFlagsSet(t *testing.T) {
	t.Run("UsernameOnly", func(t *testing.T) {
		assert.False(t, UserFlagsSet(newCommandContext(t, UsersModCommand, "--username=jane"), "username", "force"))
	})
	t.Run("OtherFlag", func(t *testing.T) {
		assert.True(t, UserFlagsSet(newCommandContext(t, UsersModCommand, "--username=jane", "--name=Jane"), "username", "force"))
	})
	t.Run("Alias", func(t *testing.T) {
		assert.True(t, UserFlagsSet(newCommandContext(t, UsersModCommand, "-n", "Jane"), "username"))
	})
	t.Run("None", func(t *testing.T) {
		assert.False(t, UserFlagsSet(newCommandContext(t, UsersModCommand), "username"))
	})
}

func TestUserUsernameFlag(t *testing.T) {
	f := UserUsernameFlag()
	assert.Equal(t, "username", f.Name)
	assert.Empty(t, f.Aliases)
	assert.Equal(t, UserUsernameUsage, f.Usage)
}

func TestWarnRenamedFolder(t *testing.T) {
	t.Run("PinnedUploadPath", func(t *testing.T) {
		logs := captureCommandLog(t)
		m := &entity.User{UserUID: "uqxetse3cy5eo9z2", UserName: "jane-new", UploadPath: "users/jane"}
		assert.True(t, WarnRenamedFolder(m, "jane"))
		assert.Equal(t, []string{"user 'jane-new' keeps the folder users/jane, assign another base path to new accounts with the handle 'jane'"}, logMessages(logs, logrus.WarnLevel))
	})
	t.Run("PinnedBasePath", func(t *testing.T) {
		assert.True(t, WarnRenamedFolder(&entity.User{UserUID: "uqxetse3cy5eo9z2", UserName: "jane-new", BasePath: "users/jane"}, "jane"))
	})
	t.Run("OtherFolder", func(t *testing.T) {
		assert.False(t, WarnRenamedFolder(&entity.User{UserUID: "uqxetse3cy5eo9z2", UserName: "jane-new", BasePath: "shared/jane"}, "jane"))
		assert.False(t, WarnRenamedFolder(&entity.User{UserUID: "uqxetse3cy5eo9z2", UserName: "jane-new"}, "jane"))
	})
	t.Run("SameHandle", func(t *testing.T) {
		assert.True(t, WarnRenamedFolder(&entity.User{UserUID: "uqxetse3cy5eo9z2", UserName: "jane@new.example", UploadPath: "~"}, "jane@old.example"))
	})
	t.Run("UploadSubfolder", func(t *testing.T) {
		assert.True(t, WarnRenamedFolder(&entity.User{UserUID: "uqxetse3cy5eo9z2", UserName: "jane-new", UploadPath: "users/jane/inbox"}, "jane"))
	})
	t.Run("Unrestricted", func(t *testing.T) {
		assert.False(t, WarnRenamedFolder(&entity.User{UserUID: "uqxetse3cy5eo9z2", UserName: "jane@new.example"}, "jane@old.example"))
	})
	t.Run("NoHandle", func(t *testing.T) {
		assert.False(t, WarnRenamedFolder(&entity.User{UserUID: "uqxetse3cy5eo9z2", UserName: "jane", BasePath: "users/+uqxetse3cy5eo9z2"}, ".jane"))
	})
	t.Run("NormalizedHandle", func(t *testing.T) {
		assert.True(t, WarnRenamedFolder(&entity.User{UserUID: "uqxetse3cy5eo9z2", UserName: "pete", UploadPath: "users/pe.te"}, "Pe$te"))
	})
	t.Run("CaseOnly", func(t *testing.T) {
		assert.False(t, WarnRenamedFolder(&entity.User{UserUID: "uqxetse3cy5eo9z2", UserName: "jane", UploadPath: "~"}, "Jane"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, WarnRenamedFolder(nil, "jane"))
		assert.False(t, WarnRenamedFolder(&entity.User{UserName: "jane-new", BasePath: "users/jane"}, ""))
	})
}
