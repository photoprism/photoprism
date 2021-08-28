package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordMap_Get(t *testing.T) {
	t.Run("get existing password", func(t *testing.T) {
		r := PasswordFixtures.Get("alice")
		assert.Equal(t, "uqxetse3cy5eo9z2", r.UID)
		assert.IsType(t, Password{}, r)
	})
	t.Run("get not existing password", func(t *testing.T) {
		r := PasswordFixtures.Get("monstera")
		assert.Equal(t, "", r.UID)
		assert.IsType(t, Password{}, r)
	})
}

func TestPasswordMap_Pointer(t *testing.T) {
	t.Run("get existing password", func(t *testing.T) {
		r := PasswordFixtures.Pointer("alice")
		assert.Equal(t, "uqxetse3cy5eo9z2", r.UID)
		assert.IsType(t, &Password{}, r)
	})
	t.Run("get not existing password", func(t *testing.T) {
		r := PasswordFixtures.Pointer("monstera")
		assert.Equal(t, "", r.UID)
		assert.IsType(t, &Password{}, r)
	})
}
