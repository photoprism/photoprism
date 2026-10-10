//go:build integration

package photoprism

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestIndex_Insta360LensNotVideo verifies that a capture file that is not a video is reported once
// and not read with ExifTool, while the capture stays stacked.
func TestIndex_Insta360LensNotVideo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cases := []struct {
		name  string
		left  string
		right string
		warn  string
		index bool
	}{
		{"TextRightLens", "mp4", "text", insta360StackRight, true},
		{"TextLeftLens", "text", "mp4", insta360StackLeft, false},
		{"Mp4Pair", "mp4", "mp4", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			folder := "insta360lens" + strings.ToLower(tc.name)
			cfg := newInsta360StackConfig(t, folder, false)
			hook := newSequenceLogHook(t)
			dir := filepath.Join(cfg.OriginalsPath(), folder)

			for name, content := range map[string]string{insta360StackLeft: tc.left, insta360StackRight: tc.right} {
				if content == "mp4" {
					writeInsta360StackMedia(t, cfg, dir, name)
				} else {
					writeInsta360LensContent(t, cfg, filepath.Join(dir, name), content)
				}
			}

			indexInsta360StackFolder(cfg, folder, false, true)

			// A main lens without a preview fails the group.
			if owners := insta360StackOwners(t, folder); tc.index {
				require.Contains(t, owners, insta360StackLeft)
				require.Contains(t, owners, insta360StackRight)
				assert.Equal(t, owners[insta360StackLeft].ID, owners[insta360StackRight].ID)
			} else {
				assert.Empty(t, owners)
			}

			var warnings []string
			for _, entry := range hook.AllEntries() {
				if entry.Level == logrus.WarnLevel && strings.Contains(entry.Message, "is not an MP4 or QuickTime video") {
					warnings = append(warnings, entry.Message)
				}
			}

			if tc.warn == "" {
				assert.Empty(t, warnings)
			} else {
				require.Len(t, warnings, 1)
				assert.Contains(t, warnings[0], folder+"/"+tc.warn)
			}

			if !cfg.ExifToolEnabled() {
				return
			}

			for _, name := range []string{insta360StackLeft, insta360StackRight} {
				f, err := NewMediaFile(filepath.Join(dir, name))
				require.NoError(t, err)
				jsonName, err := f.ExifToolJsonName()
				require.NoError(t, err)
				assert.Equal(t, name != tc.warn && tc.index, fs.FileExists(jsonName), name)
			}
		})
	}
}

// TestImport_Insta360LensNotVideo verifies that an imported capture file that is not a video is reported
// and not read with ExifTool after it was renamed.
func TestImport_Insta360LensNotVideo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	folder := "insta360importlens"
	cfg := newInsta360StackConfig(t, folder, false)
	hook := newSequenceLogHook(t)
	importDir := filepath.Join(cfg.ImportPath(), folder)

	writeInsta360StackMedia(t, cfg, importDir, insta360StackLeft)
	writeInsta360LensContent(t, cfg, filepath.Join(importDir, insta360StackRight), "text")

	convert := NewConvert(cfg)
	NewImport(cfg, NewIndex(cfg, convert, NewFiles(), NewPhotos()), convert).Start(ImportOptionsMove(importDir, folder))

	var right entity.File
	require.NoError(t, entity.UnscopedDb().First(&right, "file_root = ? AND original_name = ?", entity.RootOriginals, insta360StackRight).Error)
	require.NotEqual(t, insta360StackRight, filepath.Base(right.FileName))

	var warnings []string
	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.WarnLevel && strings.Contains(entry.Message, "is not an MP4 or QuickTime video") {
			warnings = append(warnings, entry.Message)
		}
	}

	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "import: "+right.FileName)

	if !cfg.ExifToolEnabled() {
		return
	}

	f, err := NewMediaFile(filepath.Join(cfg.OriginalsPath(), right.FileName))
	require.NoError(t, err)
	jsonName, err := f.ExifToolJsonName()
	require.NoError(t, err)
	assert.NoFileExists(t, jsonName)
}
