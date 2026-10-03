package authn

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/clean"

	"github.com/stretchr/testify/assert"
)

func TestUsername(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		name := ""
		if s, err := Username(name); err != nil {
			assert.ErrorIs(t, err, ErrEmpty)
		} else {
			assert.Equal(t, name, s)
		}
	})
	t.Run("TooLong", func(t *testing.T) {
		name := "teeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeest"
		if s, err := Username(name); err != nil {
			assert.ErrorIs(t, err, ErrTooLong)
		} else {
			assert.Equal(t, name, s)
		}
	})
	t.Run("Admin", func(t *testing.T) {
		name := "admin"
		if s, err := Username(name); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "admin", s)
		}
	})
	t.Run("Uppercase", func(t *testing.T) {
		name := "ADMIN"
		if s, err := Username(name); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("Uppercase: %s -> %s", name, s)
			assert.Equal(t, "admin", s)
		}
	})
	t.Run("Visitor", func(t *testing.T) {
		name := "Visitor"
		if s, err := Username(name); err != nil {
			assert.ErrorIs(t, err, ErrReserved)
			assert.Equal(t, "visitor", s)
		} else {
			assert.Equal(t, "visitor", s)
		}
	})
	t.Run("Asterisk", func(t *testing.T) {
		name := "*"
		if s, err := Username(name); err != nil {
			assert.ErrorIs(t, err, ErrInvalid)
		} else {
			assert.Equal(t, name, s)
		}
	})
	t.Run("ClientUID", func(t *testing.T) {
		name := "cs6sg6beu8nm9e6t"
		if s, err := Username(name); err != nil {
			assert.ErrorIs(t, err, ErrReserved)
		} else {
			assert.Equal(t, name, s)
		}
	})
}

func TestUsernameRejectsFieldSeparator(t *testing.T) {
	// clean.Username removes it, and this gate refuses any name it had to alter. The rule is
	// this function's, not the package's: clean.Handle folds the character instead.
	s, err := Username("ok" + string(clean.FieldSep) + "granted")

	assert.Equal(t, ErrInvalid, err)
	assert.NotContains(t, s, string(clean.FieldSep))
}
