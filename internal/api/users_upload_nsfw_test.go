package api

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/media"
)

// stubNSFW installs a detector result for the duration of a test.
func stubNSFW(t *testing.T, results []nsfw.Result, err error) {
	t.Helper()

	vision.SetNSFWFunc(func(vision.Files, media.Src) ([]nsfw.Result, error) {
		return results, err
	})
	previousPreview := nsfwUploadPreview
	nsfwUploadPreview = func(fileName string) (string, func(), error) {
		return fileName, func() {}, nil
	}

	t.Cleanup(func() {
		vision.SetNSFWFunc(nil)
		nsfwUploadPreview = previousPreview
	})
}

// TestUploadScreeningPreview verifies sidecars are not sent to the image detector.
func TestUploadScreeningPreview(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "photo.xmp")
	require.NoError(t, os.WriteFile(fileName, []byte("<x:xmpmeta/>"), fs.ModeFile))
	preview, cleanup, err := uploadScreeningPreview(fileName)
	require.NoError(t, err)
	assert.Empty(t, preview)
	require.NotNil(t, cleanup)
	cleanup()
}

// TestNsfwUploadStatus verifies that screening distinguishes unsafe and unavailable files.
func TestNsfwUploadStatus(t *testing.T) {
	t.Run("DisabledDetectorIsAdmitted", func(t *testing.T) {
		previous := vision.Config
		vision.Config = &vision.ConfigValues{Models: vision.Models{{Type: vision.ModelTypeNsfw, Disabled: true}}}
		t.Cleanup(func() { vision.Config = previous })
		assert.Equal(t, nsfw.StatusSafe, nsfwUploadStatus("unscreened.jpg"))
	})
	t.Run("SafeIsAdmitted", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{nsfw.NewResult(0.01, nsfw.DefaultThreshold)}, nil)
		assert.Equal(t, nsfw.StatusSafe, nsfwUploadStatus("holiday.jpg"))
	})
	t.Run("UnsafeIsRejected", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{nsfw.NewResult(0.99, nsfw.DefaultThreshold)}, nil)
		assert.Equal(t, nsfw.StatusUnsafe, nsfwUploadStatus("offensive.jpg"))
	})
	t.Run("DetectorErrorIsRejected", func(t *testing.T) {
		stubNSFW(t, nil, errors.New("inference failed"))
		assert.Equal(t, nsfw.StatusUnavailable, nsfwUploadStatus("unreadable.jpg"))
	})
	t.Run("DetectorUnavailableIsRejected", func(t *testing.T) {
		stubNSFW(t, nil, fmt.Errorf("%w: model is missing", nsfw.ErrDetectorUnavailable))
		assert.Equal(t, nsfw.StatusUnavailable, nsfwUploadStatus("unscreened.jpg"))
	})
	t.Run("NoResultIsRejected", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{}, nil)
		assert.Equal(t, nsfw.StatusUnavailable, nsfwUploadStatus("empty.jpg"))
	})
	t.Run("UndecidedResultIsRejected", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{nsfw.Unavailable("thumbnail is missing")}, nil)
		assert.Equal(t, nsfw.StatusUnavailable, nsfwUploadStatus("undecided.jpg"))
	})
	// The zero value must behave like any other undecided result.
	t.Run("ZeroResultIsRejected", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{{}}, nil)
		assert.Equal(t, nsfw.StatusUnavailable, nsfwUploadStatus("zero.jpg"))
	})
	// Screening that was never configured is the operator's choice, not a fault. Rejecting
	// here would delete every upload on an instance whose vision.yml disables the model.
	t.Run("NotConfiguredIsAdmitted", func(t *testing.T) {
		stubNSFW(t, nil, fmt.Errorf("%w: missing nsfw model", nsfw.ErrNotConfigured))
		assert.Equal(t, nsfw.StatusSafe, nsfwUploadStatus("unscreened.jpg"))
	})
	t.Run("PreviewFailureIsRejected", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{nsfw.NewResult(0.01, nsfw.DefaultThreshold)}, nil)
		nsfwUploadPreview = func(string) (string, func(), error) {
			return "", nil, errors.New("preview failed")
		}
		assert.Equal(t, nsfw.StatusUnavailable, nsfwUploadStatus("camera.raw"))
	})
}

// TestRemoveScreenedUploads verifies rejected batches are removed before returning an error.
func TestRemoveScreenedUploads(t *testing.T) {
	first := filepath.Join(t.TempDir(), "first.jpg")
	second := filepath.Join(t.TempDir(), "second.jpg")
	require.NoError(t, os.WriteFile(first, []byte("first"), fs.ModeFile))
	require.NoError(t, os.WriteFile(second, []byte("second"), fs.ModeFile))

	removeScreenedUploads([]string{first, second})
	assert.NoFileExists(t, first)
	assert.NoFileExists(t, second)
}

// TestNsfwUploadError verifies unsafe content and detector failure use distinct responses.
func TestNsfwUploadError(t *testing.T) {
	code, message := nsfwUploadError(nsfw.StatusUnsafe)
	assert.Equal(t, 403, code)
	assert.Equal(t, i18n.ErrOffensiveUpload, message)

	code, message = nsfwUploadError(nsfw.StatusUnavailable)
	assert.Equal(t, 503, code)
	assert.Equal(t, i18n.ErrContentScreeningUnavailable, message)
}
