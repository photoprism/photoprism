package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrant_Allow(t *testing.T) {
	t.Run("ViewAllDownload", func(t *testing.T) {
		assert.True(t, GrantViewAll.Allow(ActionView))
	})
	t.Run("ViewAllShare", func(t *testing.T) {
		assert.False(t, GrantViewAll.Allow(ActionShare))
	})
	t.Run("UnknownAction", func(t *testing.T) {
		assert.False(t, GrantViewAll.Allow("lovecats"))
	})
}
