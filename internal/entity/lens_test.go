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
	t.Run("IPhoneXs", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone XS back camera 4.25mm f/1.8")
		assert.Equal(t, "apple-iphone-xs-4-25mm-f-1-8", lens.LensSlug)
		assert.Equal(t, "Apple iPhone XS 4.25mm f/1.8", lens.LensName)
		assert.Equal(t, "iPhone XS 4.25mm f/1.8", lens.LensModel)
		assert.Equal(t, "Apple", lens.LensMake)
	})
	t.Run("IPhoneTwelveMini", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone 12 mini back dual wide camera 4.2mm f/1.6")
		assert.Equal(t, "apple-iphone-12-mini-4-2mm-f-1-6", lens.LensSlug)
		assert.Equal(t, "Apple iPhone 12 mini 4.2mm f/1.6", lens.LensName)
		assert.Equal(t, "iPhone 12 mini 4.2mm f/1.6", lens.LensModel)
		assert.Equal(t, "Apple", lens.LensMake)
	})
	t.Run("IPhoneTwelveUltraWide", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone 12 back dual wide camera 1.55mm f/2.4")
		assert.Equal(t, "apple-iphone-12-1-55mm-f-2-4", lens.LensSlug)
		assert.Equal(t, "Apple iPhone 12 1.55mm f/2.4", lens.LensName)
		assert.Equal(t, "iPhone 12 1.55mm f/2.4", lens.LensModel)
		assert.Equal(t, "Apple", lens.LensMake)
	})
	t.Run("IPhoneFourteenProMax", func(t *testing.T) {
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
	t.Run("ExistingLens", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone SE")

		result := FirstOrCreateLens(lens)

		if result == nil {
			t.Fatal("result must not be nil")
		}
	})
	t.Run("NotExistingLens", func(t *testing.T) {
		lens := &Lens{}

		result := FirstOrCreateLens(lens)

		if result == nil {
			t.Fatal("result must not be nil")
		}
		assert.GreaterOrEqual(t, result.ID, uint(1))
	})
}

func TestLensUpdateMakeModel(t *testing.T) {
	t.Run("ExistingLens", func(t *testing.T) {
		setup := NewLens("", "4 38")
		lens := FirstOrCreateLens(setup)
		defer assert.NoError(t, UnscopedDb().Delete(&Lens{}, "id = ?", lens.ID).Error)
		make := "Pentax"
		model := "smc PENTAX-FA 28-105mm F3.2-4.5 AL[IF]"
		err := lens.UpdateMakeModel(make, model)
		assert.NoError(t, err)
		assert.Equal(t, CameraMakes[make], lens.LensMake)
		assert.Equal(t, model, lens.LensModel)
		assert.Equal(t, "4-38", lens.LensSlug)
		assert.Equal(t, "PENTAX smc PENTAX-FA 28-105mm F3.2-4.5 AL[IF]", lens.LensName)
	})
	t.Run("NotExistingLens", func(t *testing.T) {
		lens := NewLens("", "4 39")
		err := lens.UpdateMakeModel("Pentax", "smc PENTAX-FA 31mm F1.8 AL Limited")
		assert.Error(t, err)
	})
}
