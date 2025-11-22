package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermission_Equal(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, FullAccess.Equal("full access"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, ActionUpload.Equal("full access"))
	})
}

func TestPermission_NotEqual(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		assert.False(t, FullAccess.NotEqual("full access"))
	})
	t.Run("True", func(t *testing.T) {
		assert.True(t, ActionUpload.NotEqual("full access"))
	})
}
