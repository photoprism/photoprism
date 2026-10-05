package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/leandro-lugaresi/hub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestIndex_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	initErr := cfg.InitializeTestData()
	assert.NoError(t, initErr)

	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
	imp := NewImport(cfg, ind, convert)
	opt := ImportOptionsMove(cfg.ImportPath(), "")

	imp.Start(opt)

	indexOpt := IndexOptionsAll(cfg)
	indexOpt.Rescan = false

	found, updated := ind.Start(indexOpt)
	assert.GreaterOrEqual(t, len(found), 0)
	assert.GreaterOrEqual(t, updated, 0)

	t.Logf("index run 1: found %s", english.Plural(updated, "file", "files"))
	t.Logf("index run 1: updated %s", english.Plural(updated, "file", "files"))

	time.Sleep(time.Second)

	found, updated = ind.Start(indexOpt)
	assert.GreaterOrEqual(t, len(found), 0)
	assert.GreaterOrEqual(t, updated, 0)

	t.Logf("index run 2: found %s", english.Plural(updated, "file", "files"))
	t.Logf("index run 2: updated %s", english.Plural(updated, "file", "files"))

	time.Sleep(time.Second)

	found, updated = ind.Start(indexOpt)
	assert.GreaterOrEqual(t, len(found), 0)
	assert.GreaterOrEqual(t, updated, 0)

	t.Logf("index run 3: found %s", english.Plural(updated, "file", "files"))
	t.Logf("index run 3: updated %s", english.Plural(updated, "file", "files"))
}

func TestIndex_File(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := config.TestConfig()
	initErr := cfg.InitializeTestData()
	assert.NoError(t, initErr)

	convert := NewConvert(cfg)
	ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())

	err := ind.FileName("xxx", IndexOptionsAll(cfg))

	assert.Equal(t, IndexFailed, err.Status)
}

func TestIndex_forgetReplacedPreview(t *testing.T) {
	c := Config()
	ind := NewIndex(c, NewConvert(c), NewFiles(), NewPhotos())

	// newPreview writes a preview below dir and records its time in the file cache.
	newPreview := func(t *testing.T, dir string) *MediaFile {
		fileName := filepath.Join(dir, "zzforgetpreview", "clip.mp4.jpg")
		require.NoError(t, fs.WriteString(fileName, "preview"))
		t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(fileName)) })

		img, err := NewMediaFile(fileName)
		require.NoError(t, err)
		require.False(t, ind.files.Ignore(img.RootRelName(), img.Root(), img.ModTime(), false))
		require.True(t, ind.files.Ignore(img.RootRelName(), img.Root(), img.ModTime(), false))

		return img
	}

	t.Run("Sidecar", func(t *testing.T) {
		img := newPreview(t, c.SidecarPath())
		require.True(t, img.InSidecar())

		ind.forgetReplacedPreview(img)
		assert.False(t, ind.files.Ignore(img.RootRelName(), img.Root(), img.ModTime(), false))
	})
	t.Run("Originals", func(t *testing.T) {
		img := newPreview(t, c.OriginalsPath())
		require.False(t, img.InSidecar())

		// A preview next to the original is not rewritten by a forced conversion, so it stays cached.
		ind.forgetReplacedPreview(img)
		assert.True(t, ind.files.Ignore(img.RootRelName(), img.Root(), img.ModTime(), false))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.NotPanics(t, func() { ind.forgetReplacedPreview(nil) })
	})
}

// TestIndex_StartMissingPath verifies that the notification for a missing folder names it relative to originals.
func TestIndex_StartMissingPath(t *testing.T) {
	cfg := config.TestConfig()
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())

	s := event.Subscribe("notify.error")
	defer event.Unsubscribe(s)

	opt := IndexOptionsAll(cfg)
	opt.Path = "e2e-missing/folder"
	found, updated := ind.Start(opt)
	assert.Empty(t, found)
	assert.Equal(t, 0, updated)

	assertMissingFolderNotice(t, s.Receiver, "index: directory", "e2e-missing/folder", cfg.OriginalsPath())
}

// assertMissingFolderNotice waits for the notification with the prefix and checks the folder name it shows.
func assertMissingFolderNotice(t *testing.T, receiver <-chan hub.Message, prefix, folder, storagePath string) {
	t.Helper()
	timeout := time.After(5 * time.Second)

	for {
		select {
		case msg := <-receiver:
			message, _ := msg.Fields["message"].(string)

			if !strings.HasPrefix(message, prefix) {
				continue
			}

			assert.Contains(t, message, folder)
			assert.NotContains(t, message, storagePath)
			return
		case <-timeout:
			t.Fatal("no notification")
		}
	}
}
