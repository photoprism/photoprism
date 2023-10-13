package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLens(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		lens := NewLens("", "")
		assert.Equal(t, UnknownID, lens.LensSlug)
		assert.Equal(t, "Unknown", lens.LensName)
		assert.Equal(t, "Unknown", lens.LensModel)
		assert.Equal(t, "", lens.LensMake)
		assert.Equal(t, UnknownLens.LensSlug, lens.LensSlug)
		assert.Equal(t, &UnknownLens, lens)
	})
	t.Run("Canon", func(t *testing.T) {
		lens := NewLens("Canon", "F500-99")
		assert.Equal(t, "canon-f500-99", lens.LensSlug)
		assert.Equal(t, "Canon F500-99", lens.LensName)
		assert.Equal(t, "F500-99", lens.LensModel)
		assert.Equal(t, "Canon", lens.LensMake)
	})
	t.Run("iPhoneXS", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone XS back camera 4.25mm f/1.8")
		assert.Equal(t, "apple-iphone-xs-4-25mm-f-1-8", lens.LensSlug)
		assert.Equal(t, "Apple iPhone XS 4.25mm f/1.8", lens.LensName)
		assert.Equal(t, "iPhone XS 4.25mm f/1.8", lens.LensModel)
		assert.Equal(t, "Apple", lens.LensMake)
	})
	t.Run("iPhone12mini", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone 12 mini back dual wide camera 4.2mm f/1.6")
		assert.Equal(t, "apple-iphone-12-mini-4-2mm-f-1-6", lens.LensSlug)
		assert.Equal(t, "Apple iPhone 12 mini 4.2mm f/1.6", lens.LensName)
		assert.Equal(t, "iPhone 12 mini 4.2mm f/1.6", lens.LensModel)
		assert.Equal(t, "Apple", lens.LensMake)
	})
	t.Run("iPhone12UltraWide", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone 12 back dual wide camera 1.55mm f/2.4")
		assert.Equal(t, "apple-iphone-12-1-55mm-f-2-4", lens.LensSlug)
		assert.Equal(t, "Apple iPhone 12 1.55mm f/2.4", lens.LensName)
		assert.Equal(t, "iPhone 12 1.55mm f/2.4", lens.LensModel)
		assert.Equal(t, "Apple", lens.LensMake)
	})
	t.Run("iPhone14ProMax", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone 14 Pro Max back triple camera 9mm f/2.8")
		assert.Equal(t, "apple-iphone-14-pro-max-9mm-f-2-8", lens.LensSlug)
		assert.Equal(t, "Apple iPhone 14 Pro Max 9mm f/2.8", lens.LensName)
		assert.Equal(t, "iPhone 14 Pro Max 9mm f/2.8", lens.LensModel)
		assert.Equal(t, "Apple", lens.LensMake)
		assert.Equal(t, "apple-iphone-14-pro-max-9mm-f-2-8", lens.LensSlug)
	})
}

func TestLens_TableName(t *testing.T) {
	lens := NewLens("Canon", "F500-99")
	tableName := lens.TableName()
	assert.Equal(t, "lenses", tableName)
}

func TestLens_String(t *testing.T) {
	lens := NewLens("samsung", "F500-99")
	assert.Equal(t, "'Samsung F500-99'", lens.String())
}

func TestFirstOrCreateLens(t *testing.T) {
	t.Run("existing lens", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone SE")

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
