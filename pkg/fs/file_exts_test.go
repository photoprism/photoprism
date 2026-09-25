package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExtensions_Known(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, Extensions.Known(""))
	})
	t.Run("Jpg", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/test.jpg"))
	})
	t.Run("Jpeg", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/test.jpeg"))
	})
	t.Run("Cr2", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/.xxx/test (jpg).cr2"))
	})
	t.Run("CR2", func(t *testing.T) {
		assert.True(t, Extensions.Known("testdata/test (jpg).CR2"))
	})
	t.Run("CR5", func(t *testing.T) {
		assert.False(t, Extensions.Known("testdata/test (jpg).CR5"))
	})
	t.Run("Mp4", func(t *testing.T) {
		assert.True(t, Extensions.Known("file.mp4"))
	})
	t.Run("Mxf", func(t *testing.T) {
		assert.True(t, Extensions.Known("file.mxf"))
	})
}

func TestIsSecondaryRaw(t *testing.T) {
	t.Run("Ori", func(t *testing.T) {
		assert.True(t, IsSecondaryRaw("P1010101.ori"))
		assert.True(t, IsSecondaryRaw("/photos/P1010101.ORI"))
	})
	t.Run("Orf", func(t *testing.T) {
		// The High Res Shot composite is the camera's result and stays the main file.
		assert.False(t, IsSecondaryRaw("P1010101.orf"))
	})
	t.Run("OtherRaw", func(t *testing.T) {
		assert.False(t, IsSecondaryRaw("IMG_2567.CR2"))
		assert.False(t, IsSecondaryRaw("canon_eos_6d.dng"))
	})
	t.Run("NotRaw", func(t *testing.T) {
		assert.False(t, IsSecondaryRaw("cat.jpg"))
		assert.False(t, IsSecondaryRaw(""))
		assert.False(t, IsSecondaryRaw("noextension"))
	})
}
