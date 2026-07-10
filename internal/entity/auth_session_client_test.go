package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/authn"
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

func TestAddClientSession(t *testing.T) {
	t.Run("Anonymous", func(t *testing.T) {
		sess, err := AddClientSession("", unix.Day, "metrics", authn.GrantClientCredentials, nil)

		assert.NoError(t, err)

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

		sess, err := AddClientSession("My Client App Token", unix.Day, "metrics", authn.GrantCLI, user)

		assert.NoError(t, err)

		if sess == nil {
			t.Fatal("session must not be nil")
		}

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
