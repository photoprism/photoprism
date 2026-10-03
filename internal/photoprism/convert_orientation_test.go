package photoprism

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
)

// extractTestPreview writes the untagged embedded preview of the canon_eos_6d.dng sample to the
// given file name and returns it.
func extractTestPreview(t *testing.T, cnf *config.Config, fileName string) string {
	t.Helper()

	// #nosec G204 -- arguments are the configured ExifTool binary and a test fixture path.
	out, err := exec.Command(cnf.ExifToolBin(), "-b", "-PreviewImage", filepath.Join(cnf.SamplesPath(), "canon_eos_6d.dng")).Output()
	require.NoError(t, err)
	require.Greater(t, len(out), 512)
	require.NoError(t, os.WriteFile(fileName, out, 0o600))
	require.Empty(t, exifOrientationTag(t, cnf, fileName), "the extracted preview must have no Orientation tag")

	return fileName
}

// exifOrientationTag returns the numeric Orientation tag of the file, or an empty string if it has none.
func exifOrientationTag(t *testing.T, cnf *config.Config, fileName string) string {
	t.Helper()

	// #nosec G204 -- arguments are the configured ExifTool binary and a test file path.
	out, err := exec.Command(cnf.ExifToolBin(), "-s3", "-n", "-Orientation", fileName).Output()
	require.NoError(t, err)

	return strings.TrimSpace(string(out))
}

// setExifOrientationTag writes the Orientation tag of a test file with the given ExifTool binary.
func setExifOrientationTag(t *testing.T, cnf *config.Config, fileName, orientation string) {
	t.Helper()

	// #nosec G204 -- arguments are the configured ExifTool binary, a fixed value, and a test file path.
	require.NoError(t, exec.Command(cnf.ExifToolBin(), "-q", "-overwrite_original", "-n", "-Orientation="+orientation, fileName).Run())
}

func TestConvert_writeMissingOrientation(t *testing.T) {
	cnf := config.TestConfig()

	if !cnf.ExifToolEnabled() {
		t.Skip("ExifTool must be available")
	}

	convert := NewConvert(cnf)

	t.Run("Success", func(t *testing.T) {
		fileName := extractTestPreview(t, cnf, filepath.Join(t.TempDir(), "preview.jpg"))
		written, err := convert.writeMissingOrientation(fileName, 6)
		require.NoError(t, err)
		assert.True(t, written)
		assert.Equal(t, "6", exifOrientationTag(t, cnf, fileName))
		leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(fileName), "*_exiftool_tmp"))
		assert.Empty(t, leftovers)
	})
	t.Run("KeepsExisting", func(t *testing.T) {
		for _, existing := range []string{"1", "3"} {
			fileName := extractTestPreview(t, cnf, filepath.Join(t.TempDir(), "preview.jpg"))
			setExifOrientationTag(t, cnf, fileName, existing)
			written, err := convert.writeMissingOrientation(fileName, 8)
			require.NoError(t, err, "a preview tagged %s fails the condition, which is not an error", existing)
			assert.False(t, written)
			assert.Equal(t, existing, exifOrientationTag(t, cnf, fileName))
		}
	})
	t.Run("OutOfRange", func(t *testing.T) {
		fileName := extractTestPreview(t, cnf, filepath.Join(t.TempDir(), "preview.jpg"))
		for _, orientation := range []int{-1, 0, 1, 9} {
			written, err := convert.writeMissingOrientation(fileName, orientation)
			require.NoError(t, err)
			assert.False(t, written, "orientation %d", orientation)
		}
		assert.Empty(t, exifOrientationTag(t, cnf, fileName))
	})
	t.Run("Failure", func(t *testing.T) {
		dir := t.TempDir()
		written, err := convert.writeMissingOrientation(filepath.Join(dir, "missing.jpg"), 6)
		require.Error(t, err)
		assert.False(t, written)
		assert.Contains(t, err.Error(), "missing.jpg")
		assert.NotContains(t, err.Error(), dir)
	})
}

// useExifToolStub configures a shell script as the ExifTool binary until the test ends.
func useExifToolStub(t *testing.T, cnf *config.Config, script string) {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "exiftool")
	require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\n"+script), 0o700)) //nolint:gosec // G306: test executable

	prevBin := cnf.Options().ExifToolBin
	cnf.Options().ExifToolBin = bin
	t.Cleanup(func() { cnf.Options().ExifToolBin = prevBin })
}
