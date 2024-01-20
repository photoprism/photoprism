package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientAuthentication(t *testing.T) {
	t.Run("Anonymous", func(t *testing.T) {
		sess := NewClientAuthentication("Anonymous", UnixDay, "metrics", nil)

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

		sess := NewClientAuthentication("alice", UnixDay, "metrics", user)

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

		sess := NewClientAuthentication("alice", UnixDay, "", user)

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

		sess := NewClientAuthentication("", 0, "metrics", user)

		if sess == nil {
			t.Fatal("session must not be nil")
		}

		t.Logf("sess: %#v", sess)
	})
}

func TestAddClientAuthentication(t *testing.T) {
	t.Run("Anonymous", func(t *testing.T) {
		sess, err := AddClientAuthentication("", UnixDay, "metrics", nil)

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

		sess, err := AddClientAuthentication("My Client App Token", UnixDay, "metrics", user)

		assert.NoError(t, err)

		if sess == nil {
			t.Fatal("session must not be nil")
		}

		t.Logf("sess: %#v", sess)
	})
}
