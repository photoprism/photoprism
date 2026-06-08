package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Format(""))
	})
	t.Run("AlreadyClean", func(t *testing.T) {
		assert.Equal(t, "mp4", Format("mp4"))
	})
	t.Run("MixedCase", func(t *testing.T) {
		assert.Equal(t, "magicyuv", Format("MagicYUV"))
		assert.Equal(t, "m8rg", Format("M8RG"))
	})
	t.Run("LeadingDot", func(t *testing.T) {
		assert.Equal(t, "avi", Format(".avi"))
	})
	t.Run("TrailingPunctuation", func(t *testing.T) {
		assert.Equal(t, "avi", Format("avi,"))
		assert.Equal(t, "avi", Format("avi;"))
		assert.Equal(t, "avi", Format("avi."))
	})
	t.Run("Whitespace", func(t *testing.T) {
		assert.Equal(t, "magicyuv", Format("  magicyuv  "))
		// Only space is trimmed; tabs and newlines stay as-is.
		assert.Equal(t, "\tmagicyuv", Format("\tmagicyuv "))
	})
	t.Run("Quoted", func(t *testing.T) {
		assert.Equal(t, "mp4", Format(`"mp4"`))
		assert.Equal(t, "mp4", Format("'mp4'"))
		assert.Equal(t, "mp4", Format("`mp4`"))
	})
	t.Run("PunctuationOnly", func(t *testing.T) {
		assert.Equal(t, "", Format(".,;:"))
		assert.Equal(t, "", Format("   "))
	})
	t.Run("PreservesInnerPunctuation", func(t *testing.T) {
		// Only leading/trailing characters are trimmed.
		assert.Equal(t, "h.264", Format("H.264"))
		assert.Equal(t, "video/mp4", Format("Video/MP4"))
	})
}

func TestFormatSep(t *testing.T) {
	assert.Equal(t, ",", FormatSep)
}
