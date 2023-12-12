package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScope(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		q := Scope("")
		assert.Equal(t, "", q)
	})
	t.Run("Sanitized", func(t *testing.T) {
		q := Scope(" foo:BAR webdav   openid metrics !")
		assert.Equal(t, "foo:bar metrics openid webdav", q)
	})
	t.Run("All", func(t *testing.T) {
		q := Scope("*")
		assert.Equal(t, "*", q)
	})
}
