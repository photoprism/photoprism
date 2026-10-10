package api

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
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

// newSharedEmailUser stores an account with the given name and email address, and removes it after the test.
func newSharedEmailUser(t *testing.T, name, email string, verified bool) *entity.User {
	t.Helper()

	m := entity.NewUser()
	m.UserName = name
	m.UserEmail = email
	m.UserRole = acl.RoleGuest.String()

	if verified {
		m.VerifiedAt = entity.TimeStamp()
	}

	require.NoError(t, m.Create())
	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserDetails{}, "user_uid = ?", m.UserUID).Error)
		assert.NoError(t, entity.UnscopedDb().Delete(&entity.UserSettings{}, "user_uid = ?", m.UserUID).Error)
		assert.NoError(t, entity.UnscopedDb().Delete(&entity.User{}, "user_uid = ?", m.UserUID).Error)
	})

	return m
}

func TestAuditSharedEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	s := &entity.Session{RefID: "sess1234abcd"}
	t.Run("Unique", func(t *testing.T) {
		audit, standard := captureAuditLog(t), captureLog(t)
		m := newSharedEmailUser(t, "auditemail-unique", "auditemail-unique@example.com", false)

		AuditSharedEmail(c, s, m, true)

		assert.Empty(t, audit.AllEntries())
		assert.Empty(t, standard.AllEntries())
	})
	t.Run("Created", func(t *testing.T) {
		audit, standard := captureAuditLog(t), captureLog(t)
		m := newSharedEmailUser(t, "auditemail-created1", "auditemail-created@example.com", false)
		newSharedEmailUser(t, "auditemail-created2", "auditemail-created@example.com", false)

		AuditSharedEmail(c, s, m, true)

		require.Len(t, audit.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, audit.LastEntry().Level)
		assert.Contains(t, audit.LastEntry().Message, "session sess1234abcd › user › 'auditemail-created1' › email is also assigned to 1 other account")
		assert.Empty(t, standard.AllEntries())
	})
	t.Run("Changed", func(t *testing.T) {
		audit, standard := captureAuditLog(t), captureLog(t)
		m := newSharedEmailUser(t, "auditemail-changed1", "auditemail-changed@example.com", false)
		newSharedEmailUser(t, "auditemail-changed2", "auditemail-changed@example.com", false)

		AuditSharedEmail(c, s, m, false)

		require.Len(t, audit.AllEntries(), 1)
		assert.Equal(t, logrus.InfoLevel, audit.LastEntry().Level)
		assert.Contains(t, audit.LastEntry().Message, "email is also assigned to 1 other account")
		assert.Empty(t, standard.AllEntries())
	})
	t.Run("ChangedVerified", func(t *testing.T) {
		audit, standard := captureAuditLog(t), captureLog(t)
		m := newSharedEmailUser(t, "auditemail-verified1", "auditemail-verified@example.com", false)
		newSharedEmailUser(t, "auditemail-verified2", "auditemail-verified@example.com", true)

		AuditSharedEmail(c, s, m, false)

		require.Len(t, audit.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, audit.LastEntry().Level)
		assert.Contains(t, audit.LastEntry().Message, "email is also verified for 1 other account")
		assert.Empty(t, standard.AllEntries())
	})
	t.Run("NoSession", func(t *testing.T) {
		audit := captureAuditLog(t)
		m := newSharedEmailUser(t, "auditemail-nosession1", "auditemail-nosession@example.com", false)
		newSharedEmailUser(t, "auditemail-nosession2", "auditemail-nosession@example.com", false)

		AuditSharedEmail(c, nil, m, true)
		AuditSharedEmail(c, s, nil, true)

		assert.Empty(t, audit.AllEntries())
	})
}

func TestUserEmailChangeAllowed(t *testing.T) {
	// sessionFor returns a session of the specified user.
	sessionFor := func(m *entity.User) *entity.Session {
		s := &entity.Session{}
		s.SetUser(m)
		return s
	}
	t.Run("ManageOwn", func(t *testing.T) {
		admin := &entity.User{UserUID: "uqxetse3cy5eo9z2", UserName: "admin-test", UserRole: acl.RoleAdmin.String()}
		assert.True(t, userEmailChangeAllowed(sessionFor(admin), false))
	})
	t.Run("SuperAdmin", func(t *testing.T) {
		super := &entity.User{UserUID: "uqxetse3cy5eo9z2", UserName: "super-test", UserRole: acl.RoleGuest.String(), SuperAdmin: true}
		assert.True(t, userEmailChangeAllowed(sessionFor(super), false))
	})
	t.Run("Guest", func(t *testing.T) {
		guest := &entity.User{UserUID: "uqxc08w3d0ej2283", UserName: "guest-test", UserRole: acl.RoleGuest.String()}
		assert.False(t, userEmailChangeAllowed(sessionFor(guest), false))
	})
	t.Run("ClusterJWT", func(t *testing.T) {
		assert.True(t, userEmailChangeAllowed(&entity.Session{}, true))
	})
	t.Run("NoSession", func(t *testing.T) {
		assert.False(t, userEmailChangeAllowed(nil, true))
	})
}
