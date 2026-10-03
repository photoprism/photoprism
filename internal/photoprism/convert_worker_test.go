package photoprism

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/fs"
)

// runConvertWorker converts the given file with ConvertWorker and returns once it is done.
func runConvertWorker(convert *Convert, f *MediaFile, force bool) {
	jobs := make(chan ConvertJob, 1)
	jobs <- ConvertJob{force: force, file: f, convert: convert}
	close(jobs)
	ConvertWorker(jobs)
}

func TestConvertWorker(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cnf := config.TestConfig()

	if !cnf.ExifToolEnabled() {
		t.Skip("ExifTool must be available for the RAW embedded-preview fallback")
	}

	convert := NewConvert(cnf)

	// newCopy returns a copy of a sample tagged with the given orientation, with Darktable and
	// RawTherapee disabled and no cached ExifTool JSON.
	newCopy := func(t *testing.T, sample, name, orientation string) (f *MediaFile, jsonName, jpegName string) {
		t.Helper()

		disableRaw := cnf.Options().DisableRaw
		cnf.Options().DisableRaw = true
		fileName := filepath.Join(cnf.OriginalsPath(), name)
		t.Cleanup(func() {
			cnf.Options().DisableRaw = disableRaw
			_ = os.Remove(fileName)
		})

		require.NoError(t, fs.Copy(filepath.Join(cnf.SamplesPath(), sample), fileName, true))
		setExifOrientationTag(t, cnf, fileName, orientation)

		f, err := NewMediaFile(fileName)
		require.NoError(t, err)
		jsonName, err = f.ExifToolJsonName()
		require.NoError(t, err)
		_ = os.Remove(jsonName)
		jpegName, _ = fs.FileName(fileName, cnf.SidecarPath(), cnf.OriginalsPath(), fs.ExtJpeg)
		t.Cleanup(func() {
			_ = os.Remove(jsonName)
			_ = os.Remove(jpegName)
		})

		return f, jsonName, jpegName
	}

	// refuseNativeParser makes the native parser refuse every sample, as it does a large file.
	refuseNativeParser := func(t *testing.T) {
		maxBytes := meta.ExifMaxFileBytes
		meta.ExifMaxFileBytes = 1024
		t.Cleanup(func() { meta.ExifMaxFileBytes = maxBytes })
	}

	t.Run("RawOrientationFromExifTool", func(t *testing.T) {
		f, jsonName, jpegName := newCopy(t, "canon_eos_6d.dng", "convert-worker-raw.dng", "8")
		refuseNativeParser(t)
		runConvertWorker(convert, f, true)
		require.True(t, fs.FileExistsNotEmpty(jpegName))
		assert.Equal(t, "8", exifOrientationTag(t, cnf, jpegName))
		assert.True(t, fs.FileExists(jsonName))
	})
	t.Run("ImageOrientationFromExifTool", func(t *testing.T) {
		f, _, jpegName := newCopy(t, "example.tif", "convert-worker-image.tif", "6")
		refuseNativeParser(t)
		runConvertWorker(convert, f, true)
		require.True(t, fs.FileExistsNotEmpty(jpegName))
		// #nosec G204 -- arguments are the configured ExifTool binary and a test file path.
		out, err := exec.Command(cnf.ExifToolBin(), "-s3", "-n", "-ImageWidth", "-ImageHeight", jpegName).Output()
		require.NoError(t, err)
		assert.Equal(t, []string{"67", "100"}, strings.Fields(string(out)), "the 100x67 sample must be rotated")
	})
	t.Run("ExistingPreview", func(t *testing.T) {
		f, jsonName, jpegName := newCopy(t, "canon_eos_6d.dng", "convert-worker-existing.dng", "8")
		extractTestPreview(t, cnf, jpegName)
		runConvertWorker(convert, f, false)
		assert.False(t, fs.FileExists(jsonName), "no metadata is read for a file that is not converted")
	})
	t.Run("ForcedExistingPreview", func(t *testing.T) {
		f, _, jpegName := newCopy(t, "canon_eos_6d.dng", "convert-worker-forced.dng", "8")
		extractTestPreview(t, cnf, jpegName)
		refuseNativeParser(t)
		runConvertWorker(convert, f, true)
		require.True(t, fs.FileExistsNotEmpty(jpegName))
		assert.Equal(t, "8", exifOrientationTag(t, cnf, jpegName), "a forced conversion replaces the untagged preview")
	})
	t.Run("ExifToolFailure", func(t *testing.T) {
		f, jsonName, jpegName := newCopy(t, "canon_eos_6d.dng", "convert-worker-failure.dng", "8")
		useExifToolStub(t, cnf, fmt.Sprintf("for a; do case \"$a\" in -j) exit 1;; esac; done\nexec '%s' \"$@\"\n", cnf.ExifToolBin()))
		runConvertWorker(convert, f, true)
		assert.False(t, fs.FileExists(jsonName))
		assert.True(t, fs.FileExistsNotEmpty(jpegName), "a failed metadata export must keep the preview")
	})
	t.Run("NilJob", func(t *testing.T) {
		jobs := make(chan ConvertJob, 2)
		jobs <- ConvertJob{convert: convert}
		jobs <- ConvertJob{file: &MediaFile{}}
		close(jobs)
		assert.NotPanics(t, func() { ConvertWorker(jobs) })
	})
}
