package photoprism

import (
	iofs "io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// hexHashPattern matches a hex-encoded checksum as clean.MaskHashes detects it.
var hexHashPattern = regexp.MustCompile(`[0-9a-fA-F]{32,}`)

func TestMediaFile_GenerateThumbnails_LogNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	mf, err := NewMediaFile(filepath.Join(Config().SamplesPath(), "elephants.jpg"))
	require.NoError(t, err)

	thumbPath := t.TempDir()

	// A directory in place of each thumbnail makes writing it fail.
	for _, name := range thumb.Names {
		if size := thumb.Sizes[name]; size.Uncached() {
			continue
		} else if fileName, nameErr := size.FileName(mf.Hash(), thumbPath); nameErr == nil {
			require.NoError(t, os.MkdirAll(fileName, fs.ModeDir))
		}
	}

	hook := captureLog(t)
	hook.Reset()

	require.Error(t, mf.GenerateThumbnails(thumbPath, true))

	messages := loggedMessages(hook, logrus.InfoLevel)
	require.NotEmpty(t, messages)

	for _, msg := range messages {
		assert.Contains(t, msg, "elephants.jpg")
		assert.NotRegexp(t, hexHashPattern, msg)
	}
}

func TestConvert_FixJpeg_LogNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := Config()

	if conf.DisableImageMagick() {
		t.Skip("ImageMagick is disabled")
	}

	mf, err := NewMediaFile(filepath.Join(conf.SamplesPath(), "elephants.jpg"))
	require.NoError(t, err)

	cacheName := filepath.Join(conf.MediaFileCachePath(mf.Hash()), mf.Hash()+fs.ExtJpeg)
	_ = os.Remove(cacheName)

	t.Cleanup(func() {
		_ = os.Remove(cacheName)
	})

	hook := captureLog(t)
	hook.Reset()

	fixed, err := NewConvert(conf).FixJpeg(mf, false)
	require.NoError(t, err)
	require.NotNil(t, fixed)

	var messages []string

	for _, msg := range loggedMessages(hook, logrus.InfoLevel) {
		if strings.HasPrefix(msg, "convert: ") {
			messages = append(messages, msg)
		}
	}

	require.Len(t, messages, 2)

	assert.Contains(t, messages[0], "convert: re-encoding elephants.jpg")
	assert.Contains(t, messages[1], "convert: elephants.jpg re-encoded in")

	for _, msg := range messages {
		assert.NotRegexp(t, hexHashPattern, msg)
	}
}

func TestMediaFile_Thumbnail_ErrorChain(t *testing.T) {
	mf, err := NewMediaFile(filepath.Join(Config().SamplesPath(), "elephants.jpg"))
	require.NoError(t, err)

	// A regular file in place of the thumbnail folder makes creating its subfolders fail.
	thumbPath := filepath.Join(t.TempDir(), "thumbnails")
	require.NoError(t, os.WriteFile(thumbPath, []byte("not a folder"), fs.ModeFile))

	_, err = mf.Thumbnail(thumbPath, thumb.Tile224)
	require.Error(t, err)

	var pathErr *iofs.PathError
	assert.ErrorAs(t, err, &pathErr)
	assert.NotContains(t, clean.Error(err), thumbPath)
}
