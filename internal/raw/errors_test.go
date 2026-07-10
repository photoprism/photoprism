package raw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUntrustworthy(t *testing.T) {
	t.Run("WhiteBalance", func(t *testing.T) {
		assert.True(t, Untrustworthy("Processing...\nCannot use camera white balance.\n"))
	})
	t.Run("MatrixNotFound", func(t *testing.T) {
		assert.True(t, Untrustworthy(`image.x3f: "Foo" matrix not found!`))
	})
	t.Run("CorruptData", func(t *testing.T) {
		assert.True(t, Untrustworthy("Corrupt data near 0x1234"))
	})
	t.Run("BenignWarning", func(t *testing.T) {
		assert.False(t, Untrustworthy("Warning: sidecar file requested but not found for: image.cr2"))
	})
	t.Run("VerboseProgress", func(t *testing.T) {
		assert.False(t, Untrustworthy("AHD interpolation...\nConverting to sRGB colorspace...\n"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Untrustworthy(""))
	})
}

func TestDecoderErrors(t *testing.T) {
	assert.Equal(t, WhiteBalanceError, DecoderErrors[0])
	assert.Contains(t, DecoderErrors, "matrix not found")
	assert.Contains(t, DecoderErrors, "incorrect JPEG dimensions")
}
