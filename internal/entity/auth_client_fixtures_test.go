package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientMap_Get(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		r := ClientFixtures.Get("alice")
		assert.Equal(t, "Alice", r.ClientName)
		assert.Equal(t, "cs5gfen1bgxz7s9i", r.ClientUID)
		assert.IsType(t, Client{}, r)
	})

	t.Run("Invalid", func(t *testing.T) {
		r := ClientFixtures.Get("xxx")
		assert.Equal(t, "", r.ClientName)
		assert.Equal(t, "", r.ClientUID)
		assert.IsType(t, Client{}, r)
	})
}

func TestClientMap_Pointer(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		r := ClientFixtures.Pointer("alice")
		assert.Equal(t, "cs5gfen1bgxz7s9i", r.ClientUID)
		assert.Equal(t, "Alice", r.ClientName)
		assert.IsType(t, &Client{}, r)
	})

	t.Run("Invalid", func(t *testing.T) {
		r := ClientFixtures.Pointer("xxx")
		assert.Equal(t, "", r.ClientName)
		assert.Equal(t, "", r.ClientUID)
		assert.IsType(t, &Client{}, r)
	})
}
