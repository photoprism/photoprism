package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordMap_Get(t *testing.T) {
	t.Run("GetExistingPassword", func(t *testing.T) {
		r := PasswordFixtures.Get("alice")
		assert.Equal(t, "uqxetse3cy5eo9z2", r.UID)
		assert.IsType(t, Password{}, r)
		assert.True(t, r.Valid("Alice123!"))
		assert.False(t, r.Valid("Bobbob123!"))
	})
	t.Run("GetNotExistingPassword", func(t *testing.T) {
		r := PasswordFixtures.Get("monstera")
		assert.Equal(t, "", r.UID)
		assert.IsType(t, Password{}, r)
		assert.True(t, r.Empty())
	})
}

func TestPasswordMap_Pointer(t *testing.T) {
	t.Run("GetExistingPassword", func(t *testing.T) {
		r := PasswordFixtures.Pointer("alice")
		assert.Equal(t, "uqxetse3cy5eo9z2", r.UID)
		assert.IsType(t, &Password{}, r)
		assert.True(t, r.Valid("Alice123!"))
	})
	t.Run("GetNotExistingPassword", func(t *testing.T) {
		r := PasswordFixtures.Pointer("monstera")
		assert.Equal(t, "", r.UID)
		assert.IsType(t, &Password{}, r)
		assert.True(t, r.Empty())
	})
}

func TestCreatePasswordFixtures(t *testing.T) {
	t.Run("StoredAndValid", func(t *testing.T) {
		CreatePasswordFixtures()

		for name, f := range PasswordFixtures {
			stored := FindPassword(f.UID)

			if assert.NotNil(t, stored, name) {
				assert.True(t, stored.Valid(f.Password), name)
			}
		}
	})
}
