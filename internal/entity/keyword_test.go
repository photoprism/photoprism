package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestKeyword_FirstOrCreate(t *testing.T) {
	keyword := NewKeyword("food")
	r := keyword.FirstOrCreate()
	assert.Equal(t, "food", r.Keyword)
}
