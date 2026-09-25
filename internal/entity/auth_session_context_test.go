package entity

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newContextRequest returns a request context with the given client address and user agent.
func newContextRequest(t *testing.T, remoteAddr, userAgent string) *gin.Context {
	t.Helper()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodGet, "/originals/", nil)
	c.Request.Header.Set("User-Agent", userAgent)
	c.Request.RemoteAddr = remoteAddr

	return c
}

// countSessions returns how many rows the sessions table holds for the given id.
func countSessions(t *testing.T, id string) (n int) {
	t.Helper()
	require.NoError(t, UnscopedDb().Model(&Session{}).Where("id = ?", id).Count(&n).Error)
	return n
}

// removeSessionRow removes a session row directly, leaving the session cache as it is.
func removeSessionRow(t *testing.T, id string) {
	t.Helper()
	require.NoError(t, UnscopedDb().Exec("DELETE FROM auth_sessions WHERE id = ?", id).Error)
	require.Equal(t, 0, countSessions(t, id))
}

func TestSession_UpdateContext(t *testing.T) {
	t.Run("RemovedRow", func(t *testing.T) {
		s := NewSession(3600, 0)
		s.SetClientName("context-removed")
		require.NoError(t, s.Save())

		cached, err := FindSession(s.ID)
		require.NoError(t, err)

		removeSessionRow(t, s.ID)

		err = cached.UpdateContext(newContextRequest(t, "10.1.2.3:1234", "agent-b"))

		assert.ErrorIs(t, err, ErrSessionNotFound)
		assert.Equal(t, 0, countSessions(t, s.ID), "the session row must stay deleted")

		_, err = FindSession(s.ID)
		assert.Error(t, err, "the session must no longer be cached")
	})
	t.Run("Existing", func(t *testing.T) {
		s := NewSession(3600, 0)
		s.SetClientName("context-existing")
		require.NoError(t, s.Save())
		t.Cleanup(func() { _ = s.Delete() })

		require.NoError(t, s.UpdateContext(newContextRequest(t, "10.1.2.4:1234", "agent-c")))

		var stored Session
		require.NoError(t, UnscopedDb().Where("id = ?", s.ID).First(&stored).Error)
		assert.Equal(t, "10.1.2.4", stored.ClientIP)
		assert.Equal(t, "10.1.2.4", stored.LoginIP)
		assert.NotNil(t, stored.LoginAt)
		assert.Equal(t, "agent-c", stored.UserAgent)

		// An unchanged context writes nothing and succeeds.
		assert.NoError(t, s.UpdateContext(newContextRequest(t, "10.1.2.4:1234", "agent-c")))
	})
	t.Run("NilSession", func(t *testing.T) {
		var m *Session
		assert.ErrorIs(t, m.UpdateContext(newContextRequest(t, "10.1.2.3:1234", "agent-b")), ErrSessionNotFound)
	})
	t.Run("NilContext", func(t *testing.T) {
		assert.NoError(t, NewSession(3600, 0).UpdateContext(nil))
	})
}

func TestSession_saveContext(t *testing.T) {
	t.Run("NotPersisted", func(t *testing.T) {
		m := &Session{ID: "not-a-session-id", UserAgent: "agent-d"}
		assert.NoError(t, m.saveContext())
	})
	t.Run("Missing", func(t *testing.T) {
		m := NewSession(3600, 0)
		m.UserAgent = "agent-e"
		assert.ErrorIs(t, m.saveContext(), ErrSessionNotFound)
		assert.Equal(t, 0, countSessions(t, m.ID))
	})
	t.Run("SameValues", func(t *testing.T) {
		s := NewSession(3600, 0)
		s.SetClientName("context-same")
		require.NoError(t, s.Save())
		t.Cleanup(func() { _ = s.Delete() })

		// Writing the stored values changes no row on MariaDB, which must not read as a missing session.
		assert.NoError(t, s.saveContext())
	})
}

func TestSession_VerifyStored(t *testing.T) {
	t.Run("Stored", func(t *testing.T) {
		s := NewSession(3600, 0)
		s.SetClientName("verify-stored")
		require.NoError(t, s.Save())
		t.Cleanup(func() { _ = s.Delete() })

		assert.NoError(t, s.VerifyStored())
	})
	t.Run("RemovedRow", func(t *testing.T) {
		s := NewSession(3600, 0).SetUser(UserFixtures.Pointer("alice"))
		s.SetClientName("verify-removed")
		require.NoError(t, s.Save())

		_, err := FindSession(s.ID)
		require.NoError(t, err)
		require.True(t, PreviewToken.Has(s.ID))

		removeSessionRow(t, s.ID)

		assert.ErrorIs(t, s.VerifyStored(), ErrSessionNotFound)
		assert.False(t, PreviewToken.Has(s.ID), "the preview token must be released")

		_, err = FindSession(s.ID)
		assert.Error(t, err, "the session must no longer be cached")
	})
	t.Run("NilSession", func(t *testing.T) {
		var m *Session
		assert.ErrorIs(t, m.VerifyStored(), ErrSessionNotFound)
	})
	t.Run("InvalidID", func(t *testing.T) {
		assert.NoError(t, (&Session{ID: "not-a-session-id"}).VerifyStored())
	})
}

func TestSession_Save_Stored(t *testing.T) {
	t.Run("RemovedRow", func(t *testing.T) {
		s := NewSession(3600, 0)
		s.SetClientName("save-removed")
		require.NoError(t, s.Save())

		removeSessionRow(t, s.ID)

		s.SetUserAgent("agent-f")
		assert.ErrorIs(t, s.Save(), ErrSessionNotFound)
		assert.Equal(t, 0, countSessions(t, s.ID), "the session row must stay deleted")
	})
	t.Run("New", func(t *testing.T) {
		s := NewSession(3600, 0)
		s.SetClientName("save-new")
		require.NoError(t, s.Save())
		t.Cleanup(func() { _ = s.Delete() })

		assert.Equal(t, 1, countSessions(t, s.ID))
	})
	t.Run("ClearedField", func(t *testing.T) {
		s := NewSession(3600, 0)
		s.SetClientName("save-cleared")
		s.AuthScope = "photos"
		require.NoError(t, s.Save())
		t.Cleanup(func() { _ = s.Delete() })

		s.AuthScope = ""
		s.SessTimeout = 0
		require.NoError(t, s.Save())

		var stored Session
		require.NoError(t, UnscopedDb().Where("id = ?", s.ID).First(&stored).Error)
		assert.Equal(t, "", stored.AuthScope, "a field reset to its zero value must be stored")
		assert.Equal(t, int64(0), stored.SessTimeout)
	})
	t.Run("SessionData", func(t *testing.T) {
		s := NewSession(3600, 0)
		s.SetClientName("save-data")
		require.NoError(t, s.Save())
		t.Cleanup(func() { _ = s.Delete() })

		s.SetData(NewSessionData().SetGroups([]string{"example-group"}))
		require.NoError(t, s.Save())

		var stored Session
		require.NoError(t, UnscopedDb().Where("id = ?", s.ID).First(&stored).Error)
		assert.Equal(t, []string{"example-group"}, stored.GetData().Groups, "session data must be stored")
	})
	t.Run("UpdatedAt", func(t *testing.T) {
		s := NewSession(3600, 0)
		s.SetClientName("save-updated-at")
		require.NoError(t, s.Save())
		t.Cleanup(func() { _ = s.Delete() })

		past := s.UpdatedAt.Add(-time.Hour)
		require.NoError(t, UnscopedDb().Model(&Session{}).Where("id = ?", s.ID).UpdateColumn("updated_at", past).Error)

		s.ClientName = "save-updated-at-renamed"
		require.NoError(t, s.Save())

		var stored Session
		require.NoError(t, UnscopedDb().Where("id = ?", s.ID).First(&stored).Error)
		assert.True(t, stored.UpdatedAt.After(past), "the update must set updated_at")
	})
	t.Run("Regenerated", func(t *testing.T) {
		s := NewSession(3600, 0)
		s.SetClientName("save-regenerated")
		require.NoError(t, s.Save())

		oldID := s.ID
		s.Regenerate()
		require.NoError(t, s.Save())
		t.Cleanup(func() { _ = s.Delete() })

		assert.NotEqual(t, oldID, s.ID)
		assert.Equal(t, 0, countSessions(t, oldID))
		assert.Equal(t, 1, countSessions(t, s.ID))
	})
}
