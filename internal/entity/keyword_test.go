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

func TestFirstOrCreateKeyword(t *testing.T) {
	keyword := NewKeyword("food")
	result := FirstOrCreateKeyword(keyword)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.Keyword != keyword.Keyword {
		t.Errorf("Keyword should be the same: %s %s", result.Keyword, keyword.Keyword)
	}
}
