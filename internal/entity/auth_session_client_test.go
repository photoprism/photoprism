package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/unix"
)

func TestNewClientAuthentication(t *testing.T) {
	t.Run("Anonymous", func(t *testing.T) {
		sess := NewClientAuthentication("Anonymous", unix.Day, "metrics", authn.GrantClientCredentials, nil)

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

		sess := NewClientAuthentication("alice", unix.Day, "metrics", authn.GrantPassword, user)

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

		sess := NewClientAuthentication("alice", unix.Day, "", authn.GrantCLI, user)

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

		sess := NewClientAuthentication("", 0, "metrics", authn.GrantCLI, user)

		if sess == nil {
			t.Fatal("session must not be nil")
		}

		t.Logf("sess: %#v", sess)
	})
}

func TestAddClientAuthentication(t *testing.T) {
	t.Run("Anonymous", func(t *testing.T) {
		sess, err := AddClientAuthentication("", unix.Day, "metrics", authn.GrantClientCredentials, nil)

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

		sess, err := AddClientAuthentication("My Client App Token", unix.Day, "metrics", authn.GrantCLI, user)

		assert.NoError(t, err)

		if sess == nil {
			t.Fatal("session must not be nil")
		}

		t.Logf("sess: %#v", sess)
	})
}
