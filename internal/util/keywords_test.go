package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeywords(t *testing.T) {
	t.Run("cat", func(t *testing.T) {
		keywords := Keywords("cat")
		assert.Equal(t, []string{"cat"}, keywords)
	})
	t.Run("was", func(t *testing.T) {
		keywords := Keywords("was")
		assert.Equal(t, []string(nil), keywords)
	})
}
