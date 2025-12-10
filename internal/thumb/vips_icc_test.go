package thumb

import (
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVipsSetIccProfileForInteropIndex(t *testing.T) {
	t.Run("PreservesExistingProfile", func(t *testing.T) {
		VipsInit()

		img, err := vips.LoadImageFromFile("testdata/interop_index_srgb_icc.jpg", VipsImportParams())
		require.NoError(t, err)

		iiFull := img.GetString("exif-ifd4-InteroperabilityIndex")
		require.NotEmpty(t, iiFull)
		require.True(t, img.HasICCProfile())

		originalProfile := img.GetICCProfile()
		require.NotEmpty(t, originalProfile)

		err = vipsSetIccProfileForInteropIndex(img, "interop_index_srgb_icc.jpg")
		assert.NoError(t, err)
		assert.True(t, img.HasICCProfile())
		assert.Equal(t, originalProfile, img.GetICCProfile())
	})
	t.Run("EmbedsAdobeProfileWhenMissing", func(t *testing.T) {
		VipsInit()

		img, err := vips.LoadImageFromFile("testdata/interop_index.jpg", VipsImportParams())
		require.NoError(t, err)
		require.False(t, img.HasICCProfile(), "fixture should have no embedded ICC profile")

		err = vipsSetIccProfileForInteropIndex(img, "interop_index.jpg")
		assert.NoError(t, err)
		assert.True(t, img.HasICCProfile(), "Adobe ICC profile should be embedded based on InteropIndex R03")
		assert.NotEmpty(t, img.GetICCProfile())
	})
	t.Run("NoInteropIndexNoop", func(t *testing.T) {
		VipsInit()

		img, err := vips.LoadImageFromFile("testdata/example.jpg", VipsImportParams())
		require.NoError(t, err)

		hasICCBefore := img.HasICCProfile()
		err = vipsSetIccProfileForInteropIndex(img, "example.jpg")
		assert.NoError(t, err)
		assert.Equal(t, hasICCBefore, img.HasICCProfile())
	})
	t.Run("InteropIndexSRGB_NoEmbed", func(t *testing.T) {
		VipsInit()

		img, err := vips.LoadImageFromFile("testdata/interop_index_r98.jpg", VipsImportParams())
		require.NoError(t, err)
		require.False(t, img.HasICCProfile(), "fixture should have no embedded ICC profile")

		err = vipsSetIccProfileForInteropIndex(img, "interop_index_r98.jpg")
		assert.NoError(t, err)
		assert.False(t, img.HasICCProfile(), "sRGB interop index should remain without embedded ICC")
	})
	t.Run("InteropIndexThumb_NoEmbed", func(t *testing.T) {
		VipsInit()

		img, err := vips.LoadImageFromFile("testdata/interop_index_thm.jpg", VipsImportParams())
		require.NoError(t, err)
		require.False(t, img.HasICCProfile(), "fixture should have no embedded ICC profile")

		err = vipsSetIccProfileForInteropIndex(img, "interop_index_thm.jpg")
		assert.NoError(t, err)
		assert.False(t, img.HasICCProfile(), "THM interop index should remain without embedded ICC")
	})
}
