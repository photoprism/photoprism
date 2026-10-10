package photoprism

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

// TestIndex_SequenceStack verifies the photo names of files stacked by name with sequences stripped.
func TestIndex_SequenceStack(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Run("SequenceOnly", func(t *testing.T) {
		parent := "seqonly"
		cfg := newInsta360StackConfig(t, parent, true)
		cfg.Options().SidecarYaml = true
		hook := newSequenceLogHook(t)
		folder := parent + "/b2"

		writeInsta360StackMedia(t, cfg, filepath.Join(cfg.OriginalsPath(), folder), "(1).jpg")
		writeInsta360StackMedia(t, cfg, filepath.Join(cfg.OriginalsPath(), folder), "(1).mp4")
		writeInsta360StackMedia(t, cfg, filepath.Join(cfg.OriginalsPath(), folder), "(2).jpg")
		writeInsta360StackMedia(t, cfg, filepath.Join(cfg.OriginalsPath(), parent), "b2.jpg")
		indexInsta360StackFolder(cfg, parent, false, true)

		assert.Equal(t, map[string]int{"(1)": 2, "(2)": 1}, sequencePhotoNames(t, folder))
		assert.Equal(t, map[string]int{"b2": 1}, sequencePhotoNames(t, parent))
		assert.FileExists(t, filepath.Join(cfg.SidecarPath(), folder, "(1).yml"))
		assert.FileExists(t, filepath.Join(cfg.SidecarPath(), folder, "(2).yml"))

		for _, entry := range hook.AllEntries() {
			if entry.Level <= logrus.WarnLevel {
				assert.NotContains(t, entry.Message, "photo name is empty")
				assert.NotContains(t, entry.Message, "no file extension")
			}
		}
	})
	t.Run("NamedSequence", func(t *testing.T) {
		folder := "seqnamed"
		cfg := newInsta360StackConfig(t, folder, true)
		dir := filepath.Join(cfg.OriginalsPath(), folder)

		writeInsta360StackMedia(t, cfg, dir, "IMG_0001 (1).jpg")
		writeInsta360StackMedia(t, cfg, dir, "IMG_0001 (2).jpg")
		indexInsta360StackFolder(cfg, folder, false, true)

		assert.Equal(t, map[string]int{"IMG_0001": 2}, sequencePhotoNames(t, folder))
	})
}

// sequencePhotoNames returns the number of files per photo name in a folder.
func sequencePhotoNames(t *testing.T, folder string) map[string]int {
	t.Helper()

	var files entity.Files
	require.NoError(t, entity.UnscopedDb().Where("file_name LIKE ? AND file_root = ? AND file_sidecar = 0",
		folder+"/%", entity.RootOriginals).Find(&files).Error)

	result := make(map[string]int)

	for _, file := range files {
		if strings.Contains(strings.TrimPrefix(file.FileName, folder+"/"), "/") {
			continue
		}

		var photo entity.Photo
		require.NoError(t, entity.UnscopedDb().First(&photo, "id = ?", file.PhotoID).Error)
		assert.Equal(t, folder, photo.PhotoPath)
		result[photo.PhotoName]++
	}

	return result
}

// newSequenceLogHook captures log entries until the test ends, restoring the previous hooks afterwards.
func newSequenceLogHook(t *testing.T) *test.Hook {
	t.Helper()
	logger := logrus.StandardLogger()
	oldHooks := make(logrus.LevelHooks, len(logger.Hooks))
	for level, hooks := range logger.Hooks {
		oldHooks[level] = append([]logrus.Hook(nil), hooks...)
	}
	t.Cleanup(func() { logger.ReplaceHooks(oldHooks) })
	return test.NewGlobal()
}
