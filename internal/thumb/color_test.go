package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseColor(t *testing.T) {
	t.Run("Vips", func(t *testing.T) {
		assert.Equal(t, ColorPreserve, ParseColor("", LibVips))
		assert.Equal(t, ColorPreserve, ParseColor(ColorAuto, LibVips))
		assert.Equal(t, ColorPreserve, ParseColor(ColorSRGB, LibVips))
		assert.Equal(t, ColorPreserve, ParseColor(ColorNone, LibVips))
	})
	t.Run("LegacyImagingFallback", func(t *testing.T) {
		assert.Equal(t, ColorNone, ParseColor("", "imaging"))
		assert.Equal(t, ColorSRGB, ParseColor(ColorAuto, "imaging"))
		assert.Equal(t, ColorSRGB, ParseColor(ColorSRGB, "imaging"))
		assert.Equal(t, ColorNone, ParseColor(ColorNone, "imaging"))
	})
}
