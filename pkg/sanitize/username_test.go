package sanitize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsername(t *testing.T) {
	t.Run("Admin ", func(t *testing.T) {
		assert.Equal(t, "admin", Username("Admin "))
	})
	t.Run(" Admin ", func(t *testing.T) {
		assert.Equal(t, "admin", Username(" Admin "))
	})
	t.Run(" admin ", func(t *testing.T) {
		assert.Equal(t, "admin", Username(" admin "))
	})
}
