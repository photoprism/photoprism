package photoprism

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/proc"
)

// tempPreviewDirs returns the temporary preview directories that currently exist.
func tempPreviewDirs(t *testing.T, conf *config.Config) []string {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(conf.TempPath(), "nsfw-preview-*"))
	require.NoError(t, err)

	return matches
}

// assertPreviewLogsBaseNames checks that no recorded log line contains a path.
func assertPreviewLogsBaseNames(t *testing.T, messages []string) {
	t.Helper()

	for _, message := range messages {
		assert.NotContains(t, message, "/", "log lines must name files by their base name only")
	}
}

// TestConvert_TempPreview verifies direct images are reused, other media are converted within the
// screening budget, and temporary files never outlive the call or its cleanup.
func TestConvert_TempPreview(t *testing.T) {
	conf := config.TestConfig()
	converter := NewConvert(conf)

	t.Run("DirectImage", func(t *testing.T) {
		fileName := filepath.Join("testdata", "flash.jpg")
		mediaFile, err := NewMediaFile(fileName)
		require.NoError(t, err)
		preview, cleanup, err := converter.TempPreview(mediaFile)
		require.NoError(t, err)
		assert.Equal(t, mediaFile.FileName(), preview)
		require.NotNil(t, cleanup)
		cleanup()
	})
	t.Run("Nil", func(t *testing.T) {
		_, cleanup, err := converter.TempPreview(nil)
		require.Error(t, err)
		assert.Nil(t, cleanup)
	})
	t.Run("Raw", func(t *testing.T) {
		if !conf.RawEnabled() {
			t.Skip("RAW conversion must be enabled")
		}

		hook := captureLog(t)
		before := tempPreviewDirs(t, conf)
		mediaFile, err := NewMediaFile(filepath.Join(conf.SamplesPath(), "canon_eos_6d.dng"))
		require.NoError(t, err)

		start := time.Now()
		preview, cleanup, err := converter.TempPreview(mediaFile)
		require.NoError(t, err)
		t.Logf("raw preview created in %s", time.Since(start))

		require.NotNil(t, cleanup)
		assert.FileExists(t, preview)
		assert.NotEqual(t, mediaFile.FileName(), preview)
		assert.True(t, strings.HasPrefix(preview, conf.TempPath()))
		cleanup()
		assert.NoDirExists(t, filepath.Dir(preview))
		assert.Equal(t, before, tempPreviewDirs(t, conf))
		assertPreviewLogsBaseNames(t, loggedMessages(hook, logrus.DebugLevel))
	})
	t.Run("Video", func(t *testing.T) {
		if !conf.FFmpegEnabled() {
			t.Skip("FFmpeg must be enabled")
		}

		mediaFile, err := NewMediaFile(filepath.Join(conf.SamplesPath(), "earth.mov"))
		require.NoError(t, err)

		start := time.Now()
		preview, cleanup, err := converter.TempPreview(mediaFile)
		require.NoError(t, err)
		t.Logf("video preview created in %s", time.Since(start))

		require.NotNil(t, cleanup)
		assert.FileExists(t, preview)
		cleanup()
		assert.NoDirExists(t, filepath.Dir(preview))
	})
	t.Run("Stdout", func(t *testing.T) {
		if conf.DisableExifTool() {
			t.Skip("ExifTool must be available for the embedded preview")
		}

		origRaw := conf.Options().DisableRaw
		conf.Options().DisableRaw = true
		t.Cleanup(func() { conf.Options().DisableRaw = origRaw })

		mediaFile, err := NewMediaFile(filepath.Join(conf.SamplesPath(), "canon_eos_6d.dng"))
		require.NoError(t, err)

		cmds, _, err := converter.JpegConvertCmds(mediaFile, filepath.Join(t.TempDir(), "preview.jpg"), "")
		require.NoError(t, err)
		require.NotEmpty(t, cmds)

		// Every candidate writes the embedded preview to stdout.
		for _, cmd := range cmds {
			require.Equal(t, "exiftool", filepath.Base(cmd.Cmd.Path))
		}

		preview, cleanup, err := converter.TempPreview(mediaFile)
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		assert.FileExists(t, preview)
		cleanup()
		assert.NoDirExists(t, filepath.Dir(preview))
	})
	t.Run("BudgetExhausted", func(t *testing.T) {
		mediaFile, err := NewMediaFile(filepath.Join(conf.SamplesPath(), "canon_eos_6d.dng"))
		require.NoError(t, err)

		cmds, _, err := converter.JpegConvertCmds(mediaFile, filepath.Join(t.TempDir(), "preview.jpg"), "")
		require.NoError(t, err)
		require.Greater(t, len(cmds), 1, "the fixture must have several candidates")

		// With presets, darktable needs the converter mutex, which must be released again.
		origPresets := conf.Options().RawPresets
		conf.Options().RawPresets = true
		t.Cleanup(func() { conf.Options().RawPresets = origPresets })

		hook := captureLog(t)
		before := tempPreviewDirs(t, conf)

		preview, cleanup, err := converter.tempPreview(mediaFile, NewConvertBudget(time.Millisecond))
		require.True(t, converter.cmdMutex.TryLock(), "the converter mutex must be released")
		converter.cmdMutex.Unlock()
		require.Error(t, err)
		assert.True(t, errors.Is(err, proc.ErrTimeout))
		assert.NotContains(t, err.Error(), "/")
		assert.Empty(t, preview)
		assert.Nil(t, cleanup)
		assert.Equal(t, before, tempPreviewDirs(t, conf))

		// Only the first candidate runs; the loop stops once the budget is used up.
		warnings := loggedMessages(hook, logrus.WarnLevel)
		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], "within the time allowed")
		assert.Contains(t, warnings[0], "canon_eos_6d.dng")
		assertPreviewLogsBaseNames(t, loggedMessages(hook, logrus.DebugLevel))
	})
	t.Run("UsesPreviewBudget", func(t *testing.T) {
		origLimit := uploadPreviewLimit
		uploadPreviewLimit = time.Millisecond
		t.Cleanup(func() { uploadPreviewLimit = origLimit })

		mediaFile, err := NewMediaFile(filepath.Join(conf.SamplesPath(), "canon_eos_6d.dng"))
		require.NoError(t, err)

		_, cleanup, err := converter.TempPreview(mediaFile)
		require.Error(t, err)
		assert.True(t, errors.Is(err, proc.ErrTimeout))
		assert.Nil(t, cleanup)
	})
	t.Run("NoUsablePreview", func(t *testing.T) {
		hook := captureLog(t)
		before := tempPreviewDirs(t, conf)
		fileName := filepath.Join(t.TempDir(), "broken.dng")
		require.NoError(t, os.WriteFile(fileName, bytes.Repeat([]byte("broken"), 1024), 0o600))
		mediaFile, err := NewMediaFile(fileName)
		require.NoError(t, err)

		_, cleanup, err := converter.TempPreview(mediaFile)
		require.Error(t, err)
		assert.False(t, errors.Is(err, proc.ErrTimeout))
		assert.NotContains(t, err.Error(), "/")
		assert.Nil(t, cleanup)
		assert.Equal(t, before, tempPreviewDirs(t, conf))
		assertPreviewLogsBaseNames(t, loggedMessages(hook, logrus.DebugLevel))
	})
}

// TestConvert_PreviewBudget verifies that upload screening uses its own bound by default.
func TestConvert_PreviewBudget(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		require.Greater(t, conf.ConvertTimeout(), UploadPreviewTimeout)
		assert.Equal(t, UploadPreviewTimeout, NewConvert(conf).previewBudget())
	})
	t.Run("ConvertTimeoutDisabled", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		conf.Options().ConvertTimeout = -1
		require.Zero(t, conf.ConvertTimeout())
		assert.Equal(t, UploadPreviewTimeout, NewConvert(conf).previewBudget())
	})
}

// TestUploadPreviewBudget verifies that the screening budget never exceeds either bound.
func TestUploadPreviewBudget(t *testing.T) {
	t.Run("SmallerConvertTimeout", func(t *testing.T) {
		assert.Equal(t, 30*time.Second, uploadPreviewBudget(30*time.Second, time.Minute))
	})
	t.Run("LargerConvertTimeout", func(t *testing.T) {
		assert.Equal(t, time.Minute, uploadPreviewBudget(10*time.Minute, time.Minute))
	})
	t.Run("EqualConvertTimeout", func(t *testing.T) {
		assert.Equal(t, time.Minute, uploadPreviewBudget(time.Minute, time.Minute))
	})
	t.Run("ConvertTimeoutDisabled", func(t *testing.T) {
		assert.Equal(t, time.Minute, uploadPreviewBudget(0, time.Minute))
	})
	t.Run("NegativeConvertTimeout", func(t *testing.T) {
		assert.Equal(t, time.Minute, uploadPreviewBudget(-time.Second, time.Minute))
	})
}

// testPreviewCmds returns conversion candidates running the given shell scripts.
func testPreviewCmds(scripts ...string) ConvertCmds {
	result := NewConvertCmds()

	for _, script := range scripts {
		result = append(result, NewConvertCmd(exec.Command("/bin/sh", "-c", script))) //nolint:gosec // test scripts
	}

	return result
}

// TestPreviewFromCmds verifies that each candidate is judged on its own output within the budget.
func TestPreviewFromCmds(t *testing.T) {
	jpegName, err := filepath.Abs(filepath.Join("testdata", "flash.jpg"))
	require.NoError(t, err)

	t.Run("Stdout", func(t *testing.T) {
		previewName := filepath.Join(t.TempDir(), "preview.jpg")
		cmds := testPreviewCmds("cat '" + jpegName + "'")
		assert.True(t, previewFromCmds(cmds, previewName, "test.raw", NewConvertBudget(time.Minute), nil))
		assert.FileExists(t, previewName)
	})
	t.Run("RemovesPartialOutput", func(t *testing.T) {
		previewName := filepath.Join(t.TempDir(), "preview.jpg")
		cmds := testPreviewCmds(
			"printf partial > '"+previewName+"'; exit 1",
			"cat '"+jpegName+"'",
		)
		assert.True(t, previewFromCmds(cmds, previewName, "test.raw", NewConvertBudget(time.Minute), nil))
	})
	t.Run("RemovesUndecodableOutput", func(t *testing.T) {
		previewName := filepath.Join(t.TempDir(), "preview.jpg")
		cmds := testPreviewCmds(
			"printf partial > '"+previewName+"'",
			"cat '"+jpegName+"'",
		)
		assert.True(t, previewFromCmds(cmds, previewName, "test.raw", NewConvertBudget(time.Minute), nil))
	})
	t.Run("NoUsableOutput", func(t *testing.T) {
		previewName := filepath.Join(t.TempDir(), "preview.jpg")
		cmds := testPreviewCmds("printf invalid", "exit 1")
		assert.False(t, previewFromCmds(cmds, previewName, "test.raw", NewConvertBudget(time.Minute), nil))
	})
	t.Run("EmptyEnv", func(t *testing.T) {
		previewName := filepath.Join(t.TempDir(), "preview.jpg")
		cmds := testPreviewCmds("cat '" + jpegName + "'")
		assert.True(t, previewFromCmds(cmds, previewName, "test.raw", NewConvertBudget(time.Minute), nil))
		assert.NotNil(t, cmds[0].Cmd.Env, "commands must not inherit the process environment")
		assert.Empty(t, cmds[0].Cmd.Env)
	})
	t.Run("StderrRejected", func(t *testing.T) {
		previewName := filepath.Join(t.TempDir(), "preview.jpg")
		cmds := testPreviewCmds("cat '"+jpegName+"'; echo unsupported >&2", "exit 1")
		cmds[0].WithStderrRejection("unsupported")
		assert.False(t, previewFromCmds(cmds, previewName, "test.raw", NewConvertBudget(time.Minute), nil))
	})
	t.Run("BudgetExhausted", func(t *testing.T) {
		previewName := filepath.Join(t.TempDir(), "preview.jpg")
		marker := filepath.Join(t.TempDir(), "ran")
		cmds := testPreviewCmds("sleep 5", "touch '"+marker+"'; cat '"+jpegName+"'")
		budget := NewConvertBudget(200 * time.Millisecond)
		assert.False(t, previewFromCmds(cmds, previewName, "test.raw", budget, nil))
		assert.True(t, budget.Exhausted())
		assert.NoFileExists(t, marker, "no candidate may run once the budget is used up")
	})
}

// TestConvert_CmdEnv verifies that conversion commands get the configured cache and library paths.
func TestConvert_CmdEnv(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		env := NewConvert(conf).cmdEnv()
		require.Len(t, env, 2)
		assert.Equal(t, "HOME="+conf.CmdCachePath(), env[0])
		assert.Equal(t, "LD_LIBRARY_PATH="+conf.CmdLibPath(), env[1])
	})
}
