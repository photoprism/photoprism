package commands

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestDownloadName checks the file names of downloads.
func TestDownloadName(t *testing.T) {
	timeName := regexp.MustCompile(`^\d{8}_\d{6}(_\d+)?\.mp4$`)

	t.Run("Title", func(t *testing.T) {
		name := downloadName(t.TempDir(), "Summer Trip 2020", ".mp4")
		assert.False(t, timeName.MatchString(name), name)
		assert.True(t, clean.PortableName(name), name)
	})
	t.Run("Unique", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "Summer Trip.mp4"), []byte("1"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "Summer Trip_2.mp4"), []byte("2"), fs.ModeFile))
		assert.Equal(t, "Summer Trip_3.mp4", downloadName(dir, "Summer Trip", ".mp4"))
	})
	t.Run("NotPortable", func(t *testing.T) {
		for _, title := range []string{"Lofi ~ chill beats", "PHOTO~1", "a\u202eb", ""} {
			assert.Regexp(t, timeName, downloadName(t.TempDir(), title, ".mp4"), "%q", title)
		}
	})
}
