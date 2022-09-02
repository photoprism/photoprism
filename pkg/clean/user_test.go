package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	t.Run("Admin ", func(t *testing.T) {
		assert.Equal(t, "admin", Login("Admin "))
	})
	t.Run(" Admin ", func(t *testing.T) {
		assert.Equal(t, "admin", Login(" Admin "))
	})
	t.Run(" admin ", func(t *testing.T) {
		assert.Equal(t, "admin", Login(" admin "))
	})
}

func TestEmail(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, "hello@photoprism.app", Email("hello@photoprism.app"))
	})
	t.Run("Whitespace", func(t *testing.T) {
		assert.Equal(t, "hello@photoprism.app", Email(" hello@photoprism.app "))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, "", Email(" hello-photoprism "))
	})
}

func TestRole(t *testing.T) {
	t.Run("Admin ", func(t *testing.T) {
		assert.Equal(t, "admin", Role("Admin "))
	})
	t.Run(" Admin ", func(t *testing.T) {
		assert.Equal(t, "admin", Role(" Admin "))
	})
	t.Run(" admin ", func(t *testing.T) {
		assert.Equal(t, "admin", Role(" admin "))
	})
}

func TestPassword(t *testing.T) {
	t.Run("Alnum", func(t *testing.T) {
		assert.Equal(t, "fgdg5yw4y", Password("fgdg5yw4y "))
	})
	t.Run("Upper", func(t *testing.T) {
		assert.Equal(t, "AABDF24245vgfrg", Password(" AABDF24245vgfrg "))
	})
	t.Run("Special", func(t *testing.T) {
		assert.Equal(t, "!#$T#)$%I#J$I", Password("!#$T#)$%I#J$I"))
	})
}
