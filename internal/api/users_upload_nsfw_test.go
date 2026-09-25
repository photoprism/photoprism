package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/media"
)

// stubNSFW installs a detector result for the duration of a test.
func stubNSFW(t *testing.T, results []nsfw.Result, err error) {
	t.Helper()

	vision.SetNSFWUploadFunc(func(vision.Files, media.Src) ([]nsfw.Result, error) {
		return results, err
	})
	previousPreview := nsfwUploadPreview
	nsfwUploadPreview = func(fileName string) (string, func(), error) {
		return fileName, func() {}, nil
	}

	t.Cleanup(func() {
		vision.SetNSFWUploadFunc(nil)
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
	t.Run("DetectorErrorIsUnavailable", func(t *testing.T) {
		stubNSFW(t, nil, errors.New("inference failed"))
		assert.Equal(t, nsfw.StatusUnavailable, nsfwUploadStatus("unreadable.jpg"))
	})
	t.Run("DetectorFailureIsUnavailable", func(t *testing.T) {
		stubNSFW(t, nil, fmt.Errorf("%w: model is missing", nsfw.ErrDetectorUnavailable))
		assert.Equal(t, nsfw.StatusUnavailable, nsfwUploadStatus("unscreened.jpg"))
	})
	t.Run("NoResultIsUnavailable", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{}, nil)
		assert.Equal(t, nsfw.StatusUnavailable, nsfwUploadStatus("empty.jpg"))
	})
	t.Run("UndecidedResultIsUnavailable", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{nsfw.Unavailable("thumbnail is missing")}, nil)
		assert.Equal(t, nsfw.StatusUnavailable, nsfwUploadStatus("undecided.jpg"))
	})
	// The zero value must behave like any other undecided result.
	t.Run("ZeroResultIsUnavailable", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{{}}, nil)
		assert.Equal(t, nsfw.StatusUnavailable, nsfwUploadStatus("zero.jpg"))
	})
	// Screening that was never configured is the operator's choice, not a fault. Rejecting
	// here would delete every upload on an instance whose vision.yml disables the model.
	t.Run("NotConfiguredIsAdmitted", func(t *testing.T) {
		stubNSFW(t, nil, fmt.Errorf("%w: missing nsfw model", nsfw.ErrNotConfigured))
		assert.Equal(t, nsfw.StatusSafe, nsfwUploadStatus("unscreened.jpg"))
	})
	t.Run("PreviewFailureIsUnavailable", func(t *testing.T) {
		stubNSFW(t, []nsfw.Result{nsfw.NewResult(0.01, nsfw.DefaultThreshold)}, nil)
		nsfwUploadPreview = func(string) (string, func(), error) {
			return "", nil, errors.New("preview failed")
		}
		assert.Equal(t, nsfw.StatusUnavailable, nsfwUploadStatus("camera.raw"))
	})
}

// TestNsfwUploadStatusUsesUploadDetector verifies upload screening uses its own detector path.
func TestNsfwUploadStatusUsesUploadDetector(t *testing.T) {
	stubNSFW(t, []nsfw.Result{nsfw.NewResult(0.9, 0.5)}, nil)
	vision.SetNSFWFunc(func(vision.Files, media.Src) ([]nsfw.Result, error) {
		return []nsfw.Result{nsfw.NewResult(0.1, 0.5)}, nil
	})
	t.Cleanup(func() { vision.SetNSFWFunc(nil) })

	assert.Equal(t, nsfw.StatusUnsafe, nsfwUploadStatus("upload.jpg"))
}

// TestUploadUserFilesAdmitsUnavailableScreening verifies undecided uploads remain staged.
func TestUploadUserFilesAdmitsUnavailableScreening(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.Options().StoragePath = t.TempDir()
	conf.Options().UploadAllow = "jpg"
	conf.Options().UploadNSFW = false
	UploadUserFiles(router)
	authToken := AuthenticateAdmin(app, router)
	stubNSFW(t, []nsfw.Result{nsfw.Unavailable("model is missing")}, nil)

	data, err := os.ReadFile(filepath.Clean("../../pkg/fs/testdata/directory/example.jpg"))
	require.NoError(t, err)
	body, contentType, err := buildMultipart(map[string][]byte{"example.jpg": data})
	require.NoError(t, err)

	const uploadToken = "nsfw-unavailable"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+entity.Admin.UserUID+"/upload/"+uploadToken, body)
	req.Header.Set("Content-Type", contentType)
	header.SetAuthorization(req, authToken)
	out := httptest.NewRecorder()
	app.ServeHTTP(out, req)

	require.Equal(t, http.StatusOK, out.Code, out.Body.String())
	uploadBase := filepath.Join(conf.UserStoragePath(entity.Admin.UserUID), "upload")
	assert.NotEmpty(t, findUploadedFilesForToken(t, uploadBase, uploadToken))
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

// TestAggregateNSFWStatus verifies unsafe results outrank unavailable results in any order.
func TestAggregateNSFWStatus(t *testing.T) {
	t.Run("UnsafeThenUnavailable", func(t *testing.T) {
		result := aggregateNSFWStatus(nsfw.StatusUnsafe, nsfw.StatusUnavailable)
		assert.Equal(t, nsfw.StatusUnsafe, result)
	})
	t.Run("UnavailableThenUnsafe", func(t *testing.T) {
		result := aggregateNSFWStatus(nsfw.StatusUnavailable, nsfw.StatusUnsafe)
		assert.Equal(t, nsfw.StatusUnsafe, result)
	})
	t.Run("SafeThenUnavailable", func(t *testing.T) {
		result := aggregateNSFWStatus(nsfw.StatusSafe, nsfw.StatusUnavailable)
		assert.Equal(t, nsfw.StatusUnavailable, result)
	})
	t.Run("SafeThenSafe", func(t *testing.T) {
		result := aggregateNSFWStatus(nsfw.StatusSafe, nsfw.StatusSafe)
		assert.Equal(t, nsfw.StatusSafe, result)
	})
}

// TestRejectNSFWUpload verifies only an unsafe decision rejects the upload batch.
func TestRejectNSFWUpload(t *testing.T) {
	assert.True(t, rejectNSFWUpload(nsfw.StatusUnsafe))
	assert.False(t, rejectNSFWUpload(nsfw.StatusUnavailable))
	assert.False(t, rejectNSFWUpload(nsfw.StatusSafe))
}
