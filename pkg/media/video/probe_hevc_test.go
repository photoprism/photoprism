package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsHEVCFile(t *testing.T) {
	t.Run("Hvc1Mov", func(t *testing.T) {
		assert.True(t, IsHEVCFile("testdata/quicktime-hvc1.mov"))
	})
	t.Run("Hvc1Heif", func(t *testing.T) {
		assert.True(t, IsHEVCFile("testdata/motion-photo.heif"))
	})
	t.Run("Avc1Mp4", func(t *testing.T) {
		assert.False(t, IsHEVCFile("testdata/mp4v-avc1.mp4"))
	})
	t.Run("Mp42Hvc1Mp4_AvcOnly", func(t *testing.T) {
		// Despite the filename, this sample carries an avc1 track only.
		assert.False(t, IsHEVCFile("testdata/mp42-hvc1.mp4"))
	})
	t.Run("NotFound", func(t *testing.T) {
		assert.False(t, IsHEVCFile("testdata/does-not-exist.mp4"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsHEVCFile(""))
	})
}

func TestIsHEVC(t *testing.T) {
	t.Run("Hvc1Mov", func(t *testing.T) {
		f := openTestFile(t, "testdata/quicktime-hvc1.mov")
		assert.True(t, IsHEVC(f))
	})
	t.Run("Avc1Mp4", func(t *testing.T) {
		f := openTestFile(t, "testdata/mp4v-avc1.mp4")
		assert.False(t, IsHEVC(f))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, IsHEVC(nil))
	})
}
