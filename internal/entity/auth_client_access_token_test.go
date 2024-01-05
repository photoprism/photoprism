package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientAccessToken(t *testing.T) {
	t.Run("Anonymous", func(t *testing.T) {
		sess := NewClientAccessToken("Anonymous", UnixDay, "metrics", nil)

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

		sess := NewClientAccessToken("alice", UnixDay, "metrics", user)

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

		sess := NewClientAccessToken("alice", UnixDay, "", user)

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

		sess := NewClientAccessToken("", 0, "metrics", user)

		if sess == nil {
			t.Fatal("session must not be nil")
		}

		t.Logf("sess: %#v", sess)
	})
}

func TestCreateClientAccessToken(t *testing.T) {
	t.Run("Anonymous", func(t *testing.T) {
		sess, err := CreateClientAccessToken("", UnixDay, "metrics", nil)

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

		sess, err := CreateClientAccessToken("My Client App Token", UnixDay, "metrics", user)

		assert.NoError(t, err)

		if sess == nil {
			t.Fatal("session must not be nil")
		}

		t.Logf("sess: %#v", sess)
	})
}
