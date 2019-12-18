package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewKeyword(t *testing.T) {
	t.Run("cat", func(t *testing.T) {
		keyword := NewKeyword("cat")
		assert.Equal(t, "cat", keyword.Keyword)
		assert.Equal(t, false, keyword.Skip)
	})
	t.Run("TABle", func(t *testing.T) {
		keyword := NewKeyword("TABle")
		assert.Equal(t, "table", keyword.Keyword)
		assert.Equal(t, false, keyword.Skip)
	})
}
