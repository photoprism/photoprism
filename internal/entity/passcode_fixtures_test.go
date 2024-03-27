package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasscodeMap_Get(t *testing.T) {
	t.Run("get existing passcode", func(t *testing.T) {
		r := PasscodeFixtures.Get("alice")
		assert.Equal(t, "uqxetse3cy5eo9z2", r.UID)
		assert.IsType(t, Passcode{}, r)
	})
	t.Run("get not existing passcode", func(t *testing.T) {
		r := PasscodeFixtures.Get("monstera")
		assert.Equal(t, "", r.UID)
		assert.IsType(t, Passcode{}, r)
	})
}

func TestPasscodeMap_Pointer(t *testing.T) {
	t.Run("get existing passcode", func(t *testing.T) {
		r := PasscodeFixtures.Pointer("alice")
		assert.Equal(t, "uqxetse3cy5eo9z2", r.UID)
		assert.IsType(t, &Passcode{}, r)
	})
	t.Run("get not existing passcode", func(t *testing.T) {
		r := PasscodeFixtures.Pointer("monstera")
		assert.Equal(t, "", r.UID)
		assert.IsType(t, &Passcode{}, r)
	})
}
