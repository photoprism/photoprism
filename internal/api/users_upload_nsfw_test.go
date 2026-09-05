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

// TestNsfwRejectsUpload verifies that screening requires an explicit safe decision.
func TestNsfwRejectsUpload(t *testing.T) {
	t.Run("DisabledDetectorIsAdmitted", func(t *testing.T) {
		previous := vision.Config
		vision.Config = &vision.ConfigValues{Models: vision.Models{{Type: vision.ModelTypeNsfw, Disabled: true}}}
		t.Cleanup(func() { vision.Config = previous })
		assert.False(t, nsfwRejectsUpload("unscreened.jpg"))
	})
	t.Run("SafeIsAdmitted", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{nsfw.NewResult(0.01, nsfw.DefaultThreshold)}, nil)
		assert.False(t, nsfwRejectsUpload("holiday.jpg"))
	})
	t.Run("UnsafeIsRejected", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{nsfw.NewResult(0.99, nsfw.DefaultThreshold)}, nil)
		assert.True(t, nsfwRejectsUpload("offensive.jpg"))
	})
	t.Run("DetectorErrorIsRejected", func(t *testing.T) {
		stubNSFW(t, nil, errors.New("inference failed"))
		assert.True(t, nsfwRejectsUpload("unreadable.jpg"))
	})
	t.Run("DetectorUnavailableIsRejected", func(t *testing.T) {
		stubNSFW(t, nil, fmt.Errorf("%w: model is missing", nsfw.ErrDetectorUnavailable))
		assert.True(t, nsfwRejectsUpload("unscreened.jpg"))
	})
	t.Run("NoResultIsRejected", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{}, nil)
		assert.True(t, nsfwRejectsUpload("empty.jpg"))
	})
	t.Run("UndecidedResultIsRejected", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{nsfw.Unavailable("thumbnail is missing")}, nil)
		assert.True(t, nsfwRejectsUpload("undecided.jpg"))
	})
	// The zero value must behave like any other undecided result, since that is what an
	// unfilled batch element used to be.
	t.Run("ZeroResultIsRejected", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{{}}, nil)
		assert.True(t, nsfwRejectsUpload("zero.jpg"))
	})
	// Screening that was never configured is the operator's choice, not a fault. Rejecting
	// here would delete every upload on an instance whose vision.yml disables the model.
	t.Run("NotConfiguredIsAdmitted", func(t *testing.T) {
		stubNSFW(t, nil, fmt.Errorf("%w: missing nsfw model", nsfw.ErrNotConfigured))
		assert.False(t, nsfwRejectsUpload("unscreened.jpg"))
	})
	t.Run("PreviewFailureIsRejected", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{nsfw.NewResult(0.01, nsfw.DefaultThreshold)}, nil)
		nsfwUploadPreview = func(string) (string, func(), error) {
			return "", nil, errors.New("preview failed")
		}
		assert.True(t, nsfwRejectsUpload("camera.raw"))
	})
}
