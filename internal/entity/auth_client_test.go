package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	m := NewClient()
	assert.Equal(t, "", m.AuthScope)
	assert.Equal(t, m.AuthScope, m.Scope())
	m.SetScope(" metrics WEBdav!")
	assert.Equal(t, "metrics webdav", m.AuthScope)
	assert.Equal(t, m.AuthScope, m.Scope())
}

func TestFindClient(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		expected := ClientFixtures.Get("alice")

		m := FindClient("cs5gfen1bgxz7s9i")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, m.UserUID, UserFixtures.Get("alice").UserUID)
		assert.Equal(t, expected.ClientUID, m.UID())
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Bob", func(t *testing.T) {
		expected := ClientFixtures.Get("bob")

		m := FindClient("cs5gfsvbd7ejzn8m")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, m.UserUID, UserFixtures.Get("bob").UserUID)
		assert.Equal(t, expected.ClientUID, m.UID())
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Metrics", func(t *testing.T) {
		expected := ClientFixtures.Get("metrics")

		m := FindClient("cs5cpu17n6gj2qo5")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Empty(t, m.UserUID)
		assert.Equal(t, expected.ClientUID, m.UID())
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Invalid", func(t *testing.T) {
		m := FindClient("123")
		assert.Nil(t, m)
	})
}

func TestClient_User(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		alice := ClientFixtures.Get("alice")

		m := alice.User()

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, "alice", m.UserName)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.UserUID)
		assert.Equal(t, "admin", m.UserRole)

	})
	t.Run("Metrics", func(t *testing.T) {
		metrics := ClientFixtures.Get("metrics")

		m := metrics.User()

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Empty(t, m.UserName)
		assert.Empty(t, m.UserUID)
		assert.Empty(t, m.UserRole)

	})
}
func TestClient_HasPassword(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		expected := ClientFixtures.Get("alice")

		m := FindClient("cs5gfen1bgxz7s9i")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, expected.ClientUID, m.UID())
		assert.False(t, m.HasSecret("xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.False(t, m.HasSecret("aaCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.False(t, m.HasSecret(""))
		assert.True(t, m.WrongSecret("xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.True(t, m.WrongSecret("aaCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.True(t, m.WrongSecret(""))
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Metrics", func(t *testing.T) {
		expected := ClientFixtures.Get("metrics")

		m := FindClient("cs5cpu17n6gj2qo5")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, expected.ClientUID, m.UID())
		assert.True(t, m.HasSecret("xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.False(t, m.HasSecret("aaCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.False(t, m.HasSecret(""))
		assert.False(t, m.WrongSecret("xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.True(t, m.WrongSecret("aaCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.True(t, m.WrongSecret(""))
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
}
