package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/authn"
)

// newPreviewTestUser creates an active account with a preview token and deletes it when the test ends.
func newPreviewTestUser(t *testing.T, name string) *User {
	t.Helper()

	u := NewUser()
	u.UserName = name
	u.UserRole = acl.RoleAdmin.String()
	u.CanLogin = true
	require.NoError(t, u.Create())
	require.NotEmpty(t, u.PreviewToken)

	t.Cleanup(func() {
		u.RevokeSessions(nil, authn.RevokeAllSessions)
		_ = UnscopedDb().Delete(u).Error
	})

	return u
}

// newPreviewTestSession stores a session of u with the given auth provider.
func newPreviewTestSession(t *testing.T, u *User, provider authn.ProviderType) *Session {
	t.Helper()

	s := NewSession(3600, 0)
	s.SetUser(u)
	s.AuthProvider = provider.String()
	require.NoError(t, s.Create())
	t.Cleanup(func() { _ = s.Delete() })

	return s
}

func TestUser_isInactive(t *testing.T) {
	active := func() *User {
		return &User{ID: 1000, UserUID: "usg73p55zwgr1gbq", UserName: "preview-active", UserRole: acl.RoleAdmin.String()}
	}

	t.Run("Active", func(t *testing.T) {
		assert.False(t, active().isInactive())
	})
	t.Run("Deleted", func(t *testing.T) {
		u := active()
		u.DeletedAt.Time, u.DeletedAt.Valid = time.Now(), true
		assert.True(t, u.isInactive())
	})
	t.Run("ProviderNone", func(t *testing.T) {
		u := active()
		u.AuthProvider = authn.ProviderNone.String()
		assert.True(t, u.isInactive())
	})
	t.Run("RoleNone", func(t *testing.T) {
		u := active()
		u.UserRole = acl.RoleNone.String()
		assert.True(t, u.isInactive())
	})
	t.Run("Expired", func(t *testing.T) {
		u := active()
		expired := time.Now().Add(-time.Hour)
		u.ExpiresAt = &expired
		assert.True(t, u.isInactive())

		// Super admins do not expire.
		u.SuperAdmin = true
		assert.False(t, u.isInactive())
	})
	t.Run("NotRegistered", func(t *testing.T) {
		assert.False(t, (&Visitor).isInactive())
		assert.False(t, (&UnknownUser).isInactive())
		assert.False(t, (*User)(nil).isInactive())
	})
}

func TestSession_SetUser_PreviewToken(t *testing.T) {
	t.Run("Active", func(t *testing.T) {
		u := newPreviewTestUser(t, "preview-setuser-active")
		s := newPreviewTestSession(t, u, authn.ProviderLocal)

		assert.Equal(t, u.PreviewToken, PreviewToken.Get(s.ID))
	})
	t.Run("Inactive", func(t *testing.T) {
		u := newPreviewTestUser(t, "preview-setuser-inactive")
		s := newPreviewTestSession(t, u, authn.ProviderApplication)
		require.Equal(t, u.PreviewToken, PreviewToken.Get(s.ID))

		u.AuthProvider = authn.ProviderNone.String()
		s.SetUser(u)

		assert.Empty(t, PreviewToken.Get(s.ID))
	})
	t.Run("LoadedInactive", func(t *testing.T) {
		u := newPreviewTestUser(t, "preview-loaded-inactive")
		s := newPreviewTestSession(t, u, authn.ProviderApplication)
		require.NoError(t, u.Updates(Values{"auth_provider": authn.ProviderNone.String()}))
		FlushUserSessionCache(u.UserUID)
		PreviewToken.Unset(s.ID)

		// Loading the session resolves its account before the session is cached.
		found, err := FindAnySession(s.ID)
		require.NoError(t, err)
		assert.Empty(t, PreviewToken.Get(s.ID))
		assert.True(t, found.GetUser().HasProvider(authn.ProviderNone))

		// Caching the loaded session again, e.g. after a save, does not register it either.
		CacheSession(found, time.Minute)
		assert.Empty(t, PreviewToken.Get(s.ID))
	})
	t.Run("LoadedActive", func(t *testing.T) {
		u := newPreviewTestUser(t, "preview-loaded-active")
		s := newPreviewTestSession(t, u, authn.ProviderApplication)
		FlushUserSessionCache(u.UserUID)
		PreviewToken.Unset(s.ID)

		_, err := FindAnySession(s.ID)
		require.NoError(t, err)
		assert.Equal(t, u.PreviewToken, PreviewToken.Get(s.ID))
	})
	t.Run("LoadedVisitor", func(t *testing.T) {
		s := NewSession(3600, 0)
		s.SetUser(&Visitor)
		s.SetPreviewToken("")
		require.NoError(t, s.Create())
		t.Cleanup(func() { _ = s.Delete() })
		token := s.PreviewToken
		require.NotEmpty(t, token)
		FlushUserSessionCache(Visitor.UserUID)
		PreviewToken.Unset(s.ID)

		// A link visitor's own token is registered again when the session is loaded.
		_, err := FindAnySession(s.ID)
		require.NoError(t, err)
		assert.Equal(t, token, PreviewToken.Get(s.ID))
	})
}

func TestUser_RevokeSessions_PreviewToken(t *testing.T) {
	t.Run("InactiveReleasesKeptSessions", func(t *testing.T) {
		u := newPreviewTestUser(t, "preview-revoke-inactive")
		app := newPreviewTestSession(t, u, authn.ProviderApplication)

		// Another active account that happens to hold the same preview token keeps it.
		other := newPreviewTestUser(t, "preview-revoke-other")
		require.NoError(t, other.Updates(Values{"preview_token": u.PreviewToken}))
		other.PreviewToken = u.PreviewToken
		otherSess := newPreviewTestSession(t, other, authn.ProviderLocal)

		u.AuthProvider = authn.ProviderNone.String()
		u.RevokeDerivedSessions(nil)

		_, err := FindAnySession(app.ID)
		require.NoError(t, err, "app passwords are kept")
		assert.Empty(t, PreviewToken.Get(app.ID))
		assert.Equal(t, u.PreviewToken, PreviewToken.Get(otherSess.ID))
		assert.True(t, PreviewToken.HasValue(u.PreviewToken))
	})
	t.Run("ActiveKeepsAppPasswords", func(t *testing.T) {
		u := newPreviewTestUser(t, "preview-revoke-active")
		app := newPreviewTestSession(t, u, authn.ProviderApplication)

		u.RevokeDerivedSessions(nil)

		assert.Equal(t, u.PreviewToken, PreviewToken.Get(app.ID))
	})
}

func TestUser_releasePreviewTokens(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		u := newPreviewTestUser(t, "preview-release")
		first := newPreviewTestSession(t, u, authn.ProviderApplication)
		second := newPreviewTestSession(t, u, authn.ProviderLocal)
		other := newPreviewTestSession(t, newPreviewTestUser(t, "preview-release-other"), authn.ProviderLocal)

		u.releasePreviewTokens()

		assert.Empty(t, PreviewToken.Get(first.ID))
		assert.Empty(t, PreviewToken.Get(second.ID))
		assert.NotEmpty(t, PreviewToken.Get(other.ID))
	})
	t.Run("NoUID", func(t *testing.T) {
		visitor := NewSession(3600, 0)
		visitor.SetPreviewToken("")
		require.NoError(t, visitor.Create())
		t.Cleanup(func() { _ = visitor.Delete() })
		require.NotEmpty(t, PreviewToken.Get(visitor.ID))

		(&User{}).releasePreviewTokens()

		assert.NotEmpty(t, PreviewToken.Get(visitor.ID))
	})
}

func TestUser_Delete_PreviewToken(t *testing.T) {
	u := newPreviewTestUser(t, "preview-delete")
	app := newPreviewTestSession(t, u, authn.ProviderApplication)
	require.Equal(t, u.PreviewToken, PreviewToken.Get(app.ID))

	require.NoError(t, u.Delete())

	assert.Empty(t, PreviewToken.Get(app.ID))
}
