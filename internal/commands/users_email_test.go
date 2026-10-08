package commands

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
)

// captureAuditLog replaces the audit logger for the duration of a test.
func captureAuditLog(t *testing.T) *test.Hook {
	t.Helper()

	orig := event.AuditLog
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	event.AuditLog = logger

	t.Cleanup(func() { event.AuditLog = orig })

	return hook
}

// deleteUserAfterTest removes the account with the given name, including its details and settings, after the test.
func deleteUserAfterTest(t *testing.T, name string) {
	t.Cleanup(func() {
		if m := entity.FindUserByName(name); m != nil {
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Password{}, "uid = ?", m.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserDetails{}, "user_uid = ?", m.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserSettings{}, "user_uid = ?", m.UserUID).Error)
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.User{}, "user_uid = ?", m.UserUID).Error)
		}
	})
}

// sharedEmailEntry returns the audit entry that reports a shared email address, or nil if there is none.
func sharedEmailEntry(hook *test.Hook) *logrus.Entry {
	for _, entry := range hook.AllEntries() {
		if strings.Contains(entry.Message, "email is also assigned to") || strings.Contains(entry.Message, "email is also verified for") {
			return entry
		}
	}

	return nil
}

func TestAuditSharedEmail(t *testing.T) {
	t.Run("Created", func(t *testing.T) {
		audit := captureAuditLog(t)
		deleteUserAfterTest(t, "cliemail1")
		deleteUserAfterTest(t, "cliemail2")

		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--email=cliemail@example.com", "--password=test1234", "--role=guest", "cliemail1"})
		require.NoError(t, err)
		assert.Nil(t, sharedEmailEntry(audit))

		_, err = RunWithTestContext(UsersAddCommand, []string{"add", "--email=cliemail@example.com", "--password=test1234", "--role=guest", "cliemail2"})
		require.NoError(t, err)

		entry := sharedEmailEntry(audit)
		require.NotNil(t, entry)
		assert.Equal(t, logrus.WarnLevel, entry.Level)
		assert.Equal(t, "audit: cli › user 'cliemail2' › email is also assigned to 1 other account", entry.Message)
	})
	t.Run("Changed", func(t *testing.T) {
		audit := captureAuditLog(t)
		deleteUserAfterTest(t, "cliemail3")
		deleteUserAfterTest(t, "cliemail4")

		for _, name := range []string{"cliemail3", "cliemail4"} {
			_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--email=" + name + "@example.com", "--password=test1234", "--role=guest", name})
			require.NoError(t, err)
		}

		_, err := RunWithTestContext(UsersModCommand, []string{"mod", "--role=admin", "cliemail4"})
		require.NoError(t, err)
		assert.Nil(t, sharedEmailEntry(audit))

		_, err = RunWithTestContext(UsersModCommand, []string{"mod", "--email=cliemail3@example.com", "cliemail4"})
		require.NoError(t, err)

		entry := sharedEmailEntry(audit)
		require.NotNil(t, entry)
		assert.Equal(t, logrus.InfoLevel, entry.Level)
		assert.Equal(t, "audit: cli › user 'cliemail4' › email is also assigned to 1 other account", entry.Message)
	})
	t.Run("Restored", func(t *testing.T) {
		deleteUserAfterTest(t, "cliemail5")
		deleteUserAfterTest(t, "cliemail6")

		for _, name := range []string{"cliemail5", "cliemail6"} {
			_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--email=cliemail-restore@example.com", "--password=test1234", "--role=guest", name})
			require.NoError(t, err)
		}

		for _, name := range []string{"cliemail5", "cliemail6"} {
			require.NoError(t, entity.FindUserByName(name).Delete())
		}

		audit := captureAuditLog(t)

		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--restore", "--password=test1234", "cliemail5"})
		require.NoError(t, err)
		assert.Nil(t, sharedEmailEntry(audit), "the other account is still deleted")

		_, err = RunWithTestContext(UsersModCommand, []string{"mod", "--restore", "cliemail6"})
		require.NoError(t, err)

		entry := sharedEmailEntry(audit)
		require.NotNil(t, entry)
		assert.Equal(t, logrus.WarnLevel, entry.Level)
		assert.Equal(t, "audit: cli › user 'cliemail6' › email is also assigned to 1 other account", entry.Message)
		audit.Reset()

		require.NoError(t, entity.FindUserByName("cliemail5").Delete())
		_, err = RunWithTestContext(UsersAddCommand, []string{"add", "--restore", "--password=test1234", "cliemail5"})
		require.NoError(t, err)

		entry = sharedEmailEntry(audit)
		require.NotNil(t, entry)
		assert.Equal(t, "audit: cli › user 'cliemail5' › email is also assigned to 1 other account", entry.Message)
	})
	t.Run("Nil", func(t *testing.T) {
		audit := captureAuditLog(t)
		AuditSharedEmail(nil, true)
		assert.Empty(t, audit.AllEntries())
	})
}
