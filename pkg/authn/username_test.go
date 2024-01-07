package authn

import (
	"testing"

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
