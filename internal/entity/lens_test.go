package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLens(t *testing.T) {
	t.Run("name F500-99 make Canon", func(t *testing.T) {
		lens := NewLens("F500-99", "Canon")
		assert.Equal(t, "canon-f500-99", lens.LensSlug)
		assert.Equal(t, "Canon F500-99", lens.LensName)
		assert.Equal(t, "F500-99", lens.LensModel)
		assert.Equal(t, "Canon", lens.LensMake)
	})
	t.Run("name Unknown make Unknown", func(t *testing.T) {
		lens := NewLens("", "")
		assert.Equal(t, "zz", lens.LensSlug)
		assert.Equal(t, "Unknown", lens.LensName)
		assert.Equal(t, "Unknown", lens.LensModel)
		assert.Equal(t, "", lens.LensMake)
		assert.Equal(t, UnknownLens.LensSlug, lens.LensSlug)
		assert.Equal(t, &UnknownLens, lens)
	})
}

func TestLens_TableName(t *testing.T) {
	lens := NewLens("F500-99", "Canon")
	tableName := lens.TableName()
	assert.Equal(t, "lenses", tableName)
}
