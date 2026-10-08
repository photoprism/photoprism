package encode

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInputFormatWhitelist(t *testing.T) {
	t.Run("Containers", func(t *testing.T) {
		list := strings.Split(InputFormatWhitelist(), ",")
		assert.Subset(t, list, []string{"mov", "matroska", "avi", "mpegts", "gif", "apng", "webp_pipe"})
	})
	t.Run("NoPlaylistsOrSequences", func(t *testing.T) {
		list := strings.Split(InputFormatWhitelist(), ",")
		for _, name := range []string{"concat", "hls", "dash", "image2", "jpeg_pipe", "tee"} {
			assert.NotContains(t, list, name)
		}
	})
}

func TestInputArgs(t *testing.T) {
	t.Run("Jpeg", func(t *testing.T) {
		assert.Equal(t, []string{"-f", "jpeg_pipe", "-i", "/a/b.JPG"}, InputArgs("/a/b.JPG"))
	})
	t.Run("Insp", func(t *testing.T) {
		assert.Equal(t, []string{"-f", "jpeg_pipe", "-i", "/a/b.insp"}, InputArgs("/a/b.insp"))
	})
	t.Run("Insv", func(t *testing.T) {
		assert.Equal(t, []string{"-f", "mov", "-i", "/a/b.insv"}, InputArgs("/a/b.insv"))
	})
	t.Run("Lrv", func(t *testing.T) {
		assert.Equal(t, []string{"-f", "mov", "-i", "/a/b.lrv"}, InputArgs("/a/b.lrv"))
	})
	t.Run("Video", func(t *testing.T) {
		assert.Equal(t, []string{"-format_whitelist", InputFormatWhitelist(), "-i", "/a/v%d.mp4"}, InputArgs("/a/v%d.mp4"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, []string{"-format_whitelist", InputFormatWhitelist(), "-i", "SRC"}, InputArgs("SRC"))
	})
}
