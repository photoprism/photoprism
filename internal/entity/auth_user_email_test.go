package entity

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
)

// newEmailTestUser stores an account with the given name and email address, and removes it after the test.
func newEmailTestUser(t *testing.T, name, email string, verified bool) *User {
	t.Helper()

	m := NewUser()
	m.UserName = name
	m.UserEmail = email
	m.UserRole = acl.RoleGuest.String()

	if verified {
		m.VerifiedAt = TimeStamp()
	}

	require.NoError(t, m.Create())
	t.Cleanup(func() { deleteTestUser(t, m) })

	return m
}

// captureEmailLogs replaces the system and standard loggers for the duration of a test.
func captureEmailLogs(t *testing.T) (system, standard *test.Hook) {
	t.Helper()

	origSystem, origLog := event.SystemLog, log
	systemLogger, system := test.NewNullLogger()
	systemLogger.SetLevel(logrus.TraceLevel)
	standardLogger, standard := test.NewNullLogger()
	standardLogger.SetLevel(logrus.TraceLevel)
	event.SystemLog, log = systemLogger, standardLogger

	t.Cleanup(func() { event.SystemLog, log = origSystem, origLog })

	return system, standard
}

func TestUser_OtherEmailHolders(t *testing.T) {
	t.Run("Unique", func(t *testing.T) {
		m := newEmailTestUser(t, "emailholders-unique", "emailholders-unique@example.com", false)

		count, verified, err := m.OtherEmailHolders()
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.False(t, verified)
	})
	t.Run("Shared", func(t *testing.T) {
		m := newEmailTestUser(t, "emailholders-shared1", "emailholders-shared@example.com", false)
		newEmailTestUser(t, "emailholders-shared2", "emailholders-shared@example.com", false)

		count, verified, err := m.OtherEmailHolders()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.False(t, verified)
	})
	t.Run("SharedVerified", func(t *testing.T) {
		m := newEmailTestUser(t, "emailholders-verified1", "emailholders-verified@example.com", false)
		newEmailTestUser(t, "emailholders-verified2", "emailholders-verified@example.com", false)
		newEmailTestUser(t, "emailholders-verified3", "emailholders-verified@example.com", true)

		count, verified, err := m.OtherEmailHolders()
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.True(t, verified)
	})
	t.Run("DifferentCase", func(t *testing.T) {
		m := newEmailTestUser(t, "emailholders-case1", "emailholders-case@example.com", false)
		other := newEmailTestUser(t, "emailholders-case2", "emailholders-case@example.com", false)
		require.NoError(t, UnscopedDb().Model(other).Update("UserEmail", "EmailHolders-Case@Example.com").Error)

		count, verified, err := m.OtherEmailHolders()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.False(t, verified)
	})
	t.Run("DeletedHolder", func(t *testing.T) {
		m := newEmailTestUser(t, "emailholders-deleted1", "emailholders-deleted@example.com", false)
		newEmailTestUser(t, "emailholders-deleted2", "emailholders-deleted@example.com", false)
		other := newEmailTestUser(t, "emailholders-deleted3", "emailholders-deleted@example.com", true)
		require.NoError(t, UnscopedDb().Model(other).Update("DeletedAt", Now()).Error)

		count, verified, err := m.OtherEmailHolders()
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.False(t, verified)
	})
	t.Run("NoEmail", func(t *testing.T) {
		count, verified, err := (&User{UserUID: "u000000000000000"}).OtherEmailHolders()
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.False(t, verified)
	})
	t.Run("Nil", func(t *testing.T) {
		count, verified, err := (*User)(nil).OtherEmailHolders()
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.False(t, verified)
	})
}

func TestUser_ReportSharedEmail(t *testing.T) {
	t.Run("Unique", func(t *testing.T) {
		system, standard := captureEmailLogs(t)
		m := newEmailTestUser(t, "reportemail-unique", "reportemail-unique@example.com", false)

		count, _ := m.ReportSharedEmail()

		assert.Equal(t, 0, count)
		assert.Empty(t, system.AllEntries())
		assert.Empty(t, standard.AllEntries())
	})
	t.Run("Shared", func(t *testing.T) {
		system, standard := captureEmailLogs(t)
		m := newEmailTestUser(t, "reportemail-shared1", "reportemail-shared@example.com", false)
		newEmailTestUser(t, "reportemail-shared2", "reportemail-shared@example.com", false)

		count, verified := m.ReportSharedEmail()

		assert.Equal(t, 1, count)
		assert.False(t, verified)
		require.Len(t, system.AllEntries(), 1)
		assert.Equal(t, logrus.InfoLevel, system.LastEntry().Level)
		assert.Equal(t, "users: 'reportemail-shared1' › email is also assigned to 1 other account", system.LastEntry().Message)
		assert.Empty(t, standard.AllEntries())
	})
	t.Run("SharedVerified", func(t *testing.T) {
		system, standard := captureEmailLogs(t)
		m := newEmailTestUser(t, "reportemail-verified1", "reportemail-verified@example.com", false)
		newEmailTestUser(t, "reportemail-verified2", "reportemail-verified@example.com", true)
		newEmailTestUser(t, "reportemail-verified3", "reportemail-verified@example.com", false)

		count, verified := m.ReportSharedEmail()

		assert.Equal(t, 2, count)
		assert.True(t, verified)
		require.Len(t, system.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, system.LastEntry().Level)
		assert.Equal(t, "users: 'reportemail-verified1' › email is also verified for 2 other accounts", system.LastEntry().Message)
		assert.Empty(t, standard.AllEntries())
	})
}
