package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
)

func TestNewClientSession(t *testing.T) {
	t.Run("Anonymous", func(t *testing.T) {
		sess := NewClientSession("Anonymous", unix.Day, "metrics", authn.GrantClientCredentials, nil)

		if sess == nil {
			t.Fatal("session must not be nil")
		}

		t.Logf("sess: %#v", sess)
	})
	t.Run("Alice", func(t *testing.T) {
		user := FindUserByName("alice")

		if user == nil {
			t.Fatal("user must not be nil")
		}

		sess := NewClientSession("alice", unix.Day, "metrics", authn.GrantPassword, user)

		if sess == nil {
			t.Fatal("session must not be nil")
		}

		t.Logf("sess: %#v", sess)
	})
	t.Run("NoScope", func(t *testing.T) {
		user := FindUserByName("alice")

		if user == nil {
			t.Fatal("user must not be nil")
		}

		sess := NewClientSession("alice", unix.Day, "", authn.GrantCLI, user)

		if sess == nil {
			t.Fatal("session must not be nil")
		}

		t.Logf("sess: %#v", sess)
	})
	t.Run("NoLifetime", func(t *testing.T) {
		user := FindUserByName("alice")

		if user == nil {
			t.Fatal("user must not be nil")
		}

		sess := NewClientSession("", 0, "metrics", authn.GrantCLI, user)

		if sess == nil {
			t.Fatal("session must not be nil")
		}

		t.Logf("sess: %#v", sess)
	})
}

func TestNewClientSession_ReleasesTokensOnDelete(t *testing.T) {
	// Reproduces #5733: an app-password session inherits the user's preview and download
	// tokens (via SetUser) before SetAuthToken finalizes its ID. Deleting the app password
	// and the user's remaining session must release the tokens from the lookup cache so they
	// stop resolving once no active session or app password uses them.
	u := &User{ //nolint:gosec // test fixture user, not a credential
		UserUID:      rnd.GenerateUID(UserUID),
		UserName:     "app-pw-lifecycle",
		UserRole:     acl.RoleAdmin.String(),
		CanLogin:     true,
		PreviewToken: "app-pw-preview-token",
	}
	if err := u.Save(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Delete(u).Error)
	})
	// Mint an app-password session for the user, as the OAuth token handler does.
	appPw := NewClientSession("app-pw-client", unix.Day, "*", authn.GrantPassword, u)
	if err := appPw.Save(); err != nil {
		t.Fatal(err)
	}

	// The user logs in, opening a browser session that shares the same tokens.
	browser := NewSession(unix.Hour, 0)
	browser.SetUser(u)
	if err := browser.Save(); err != nil {
		t.Fatal(err)
	}

	assert.True(t, PreviewToken.HasValue("app-pw-preview-token"))

	// Deleting the app password (loaded fresh by ref ID, as the API handler does) keeps the
	// tokens valid because the browser session still uses them.
	found := FindSessionByRefID(appPw.RefID)
	if found == nil {
		t.Fatal("app password session not found by ref id")
	}
	if err := found.Delete(); err != nil {
		t.Fatal(err)
	}
	assert.True(t, PreviewToken.HasValue("app-pw-preview-token"))

	// Logging out releases the tokens once no session or app password uses them.
	if err := browser.Delete(); err != nil {
		t.Fatal(err)
	}
	assert.True(t, PreviewToken.MissingValue("app-pw-preview-token"))
	assert.NoError(t, UnscopedDb().Delete(u).Error)
}

func TestAddClientSession(t *testing.T) {
	t.Run("Anonymous", func(t *testing.T) {
		sess, err := AddClientSession("", unix.Day, "metrics", authn.GrantClientCredentials, nil)

		assert.NoError(t, err)

		if sess == nil {
			t.Fatal("session must not be nil")
		}
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Delete(sess).Error)
		})

		t.Logf("sess: %#v", sess)
	})
	t.Run("Alice", func(t *testing.T) {
		user := FindUserByName("alice")

		if user == nil {
			t.Fatal("user must not be nil")
		}

		sess, err := AddClientSession("My Client App Token", unix.Day, "metrics", authn.GrantCLI, user)

		assert.NoError(t, err)

		if sess == nil {
			t.Fatal("session must not be nil")
		}
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Delete(sess).Error)
		})

		t.Logf("sess: %#v", sess)
	})
	t.Run("NoClientIPPersistsNullLoginAt", func(t *testing.T) {
		// Sessions created without a client IP (e.g. via "photoprism auth add")
		// must persist login_at as SQL NULL, not a zero "0000-00-00" datetime
		// that strict MySQL/MariaDB sql_modes reject with Error 1292.
		// A value-typed LoginAt leaves Go's zero time.Time, which go-sql-driver/mysql serializes as the literal 0000-00-00.
		sess, err := AddClientSession("", unix.Day, "metrics", authn.GrantCLI, nil)

		assert.NoError(t, err)

		if sess == nil {
			t.Fatal("session must not be nil")
		}
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Delete(sess).Error)
		})

		assert.Nil(t, sess.LoginAt)

		var nullCount int
		if err = UnscopedDb().Table("auth_sessions").
			Where("id = ? AND login_at IS NULL", sess.ID).
			Count(&nullCount).Error; err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, nullCount, "login_at must be NULL when no client IP was set")
	})
	t.Run("ClientIPSetsLoginAt", func(t *testing.T) {
		sess := NewClientSession("", unix.Day, "metrics", authn.GrantClientCredentials, nil)
		sess.SetClientIP("203.0.113.7")

		assert.NoError(t, sess.Create())
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Delete(sess).Error)
		})

		if sess.LoginAt == nil {
			t.Fatal("login_at must be set when a client IP is present")
		}

		var nullCount int
		if err := UnscopedDb().Table("auth_sessions").
			Where("id = ? AND login_at IS NULL", sess.ID).
			Count(&nullCount).Error; err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, nullCount, "login_at must not be NULL when a client IP was set")
	})
}
