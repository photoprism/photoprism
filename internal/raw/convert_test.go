package raw

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDarktableCmd(t *testing.T) {
	t.Run("Presets", func(t *testing.T) {
		cmd, useMutex := DarktableCmd(DarktableOptions{
			Bin:      "/usr/bin/darktable-cli",
			RawName:  "image.cr2",
			JpegName: "image.cr2.jpg",
			MaxSize:  7680,
			Presets:  true,
		})
		assert.True(t, useMutex)
		s := cmd.String()
		assert.Contains(t, s, "/usr/bin/darktable-cli")
		assert.Contains(t, s, "image.cr2 image.cr2.jpg")
		assert.Contains(t, s, "--width 7680 --height 7680")
		assert.NotContains(t, s, "--apply-custom-presets")
	})
	t.Run("NoPresetsWithXmpAndDirs", func(t *testing.T) {
		cmd, useMutex := DarktableCmd(DarktableOptions{
			Bin:       "/usr/bin/darktable-cli",
			RawName:   "image.cr2",
			XmpName:   "image.cr2.xmp",
			JpegName:  "image.cr2.jpg",
			MaxSize:   1920,
			Presets:   false,
			ConfigDir: "/cfg",
			CacheDir:  "/cache",
		})
		assert.False(t, useMutex)
		s := cmd.String()
		assert.Contains(t, s, "image.cr2 image.cr2.xmp image.cr2.jpg")
		assert.Contains(t, s, "--apply-custom-presets false")
		assert.Contains(t, s, "--configdir /cfg")
		assert.Contains(t, s, "--cachedir /cache")
	})
}

func TestTherapeeCmd(t *testing.T) {
	cmd := TherapeeCmd("/usr/bin/rawtherapee-cli", "image.cr2", "image.cr2.jpg", "/assets/raw.pp3", 92)
	s := cmd.String()
	assert.Contains(t, s, "/usr/bin/rawtherapee-cli")
	assert.Contains(t, s, "-o image.cr2.jpg")
	assert.Contains(t, s, "-p /assets/raw.pp3")
	assert.Contains(t, s, "-j92")
	assert.Contains(t, s, "-c image.cr2")
}

func TestExifToolJpgFromRawCmd(t *testing.T) {
	cmd := ExifToolJpgFromRawCmd("/usr/bin/exiftool", "image.cr2")
	assert.Equal(t, []string{"/usr/bin/exiftool", "-q", "-q", "-b", "-JpgFromRaw", "image.cr2"}, cmd.Args)
}

func TestExifToolPreviewImageCmd(t *testing.T) {
	cmd := ExifToolPreviewImageCmd("/usr/bin/exiftool", "image.cr2")
	assert.True(t, strings.HasSuffix(cmd.String(), "-q -q -b -PreviewImage image.cr2"))
}
