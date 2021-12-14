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
		assert.Equal(t, UnknownID, lens.LensSlug)
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

func TestLens_String(t *testing.T) {
	lens := NewLens("F500-99", "samsung")
	assert.Equal(t, "'Samsung F500-99'", lens.String())
}

func TestFirstOrCreateLens(t *testing.T) {
	t.Run("existing lens", func(t *testing.T) {
		lens := NewLens("iPhone SE", "Apple")

		result := FirstOrCreateLens(lens)

		if result == nil {
			t.Fatal("result should not be nil")
		}
	})
	t.Run("not existing lens", func(t *testing.T) {
		lens := &Lens{}

		result := FirstOrCreateLens(lens)

		if result == nil {
			t.Fatal("result should not be nil")
		}
		assert.GreaterOrEqual(t, result.ID, uint(1))
	})
}
