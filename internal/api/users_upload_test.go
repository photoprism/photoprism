package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestUploadUserFiles(t *testing.T) {
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/upload/abc123456789", adminUid)
		// t.Logf("Request URL: %s", reqUrl)
		UploadUserFiles(router)
		r := PerformRequestWithBody(app, "POST", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("ReadOnlyMode", func(t *testing.T) {
		app, router, config := NewApiTest()
		config.Options().ReadOnly = true
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/upload/abc123456789", adminUid)
		// t.Logf("Request URL: %s", reqUrl)
		UploadUserFiles(router)
		r := PerformRequestWithBody(app, "POST", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusForbidden, r.Code)
		config.Options().ReadOnly = false
	})
	t.Run("QuotaExceeded", func(t *testing.T) {
		app, router, config := NewApiTest()
		config.Options().FilesQuota = 1
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/upload/abc123456789", adminUid)
		// t.Logf("Request URL: %s", reqUrl)
		UploadUserFiles(router)
		r := PerformRequestWithBody(app, "POST", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusInsufficientStorage, r.Code)
		config.Options().FilesQuota = 0
	})
}

func TestUploadCheckFile_AcceptsAndReducesLimit(t *testing.T) {
	dir := t.TempDir()
	// Copy a small known-good JPEG test file from pkg/fs/testdata
	src := filepath.Clean("../../pkg/fs/testdata/directory/example.jpg")
	dst := filepath.Join(dir, "example.jpg")
	b, err := os.ReadFile(src)
	if err != nil {
		t.Skipf("skip if test asset not present: %v", err)
	}
	if err := os.WriteFile(dst, b, 0o600); err != nil {
		t.Fatal(err)
	}

	orig := int64(len(b))
	rem, err := UploadCheckFile(dst, false, orig+100)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), rem)
	// file remains
	assert.FileExists(t, dst)
}

func TestUploadCheckFile_TotalLimitReachedDeletes(t *testing.T) {
	dir := t.TempDir()
	// Make a tiny file
	dst := filepath.Join(dir, "tiny.txt")
	assert.NoError(t, os.WriteFile(dst, []byte("hello"), 0o600))
	// Very small total limit (0) → should remove file and error
	_, err := UploadCheckFile(dst, false, 0)
	assert.Error(t, err)
	_, statErr := os.Stat(dst)
	assert.True(t, os.IsNotExist(statErr), "file should be removed when limit reached")
}

func TestUploadCheckFile_UnsupportedTypeDeletes(t *testing.T) {
	dir := t.TempDir()
	// Create a file with an unknown extension; should be rejected
	dst := filepath.Join(dir, "unknown.xyz")
	assert.NoError(t, os.WriteFile(dst, []byte("not-an-image"), 0o600))
	_, err := UploadCheckFile(dst, false, 1<<20)
	assert.Error(t, err)
	_, statErr := os.Stat(dst)
	assert.True(t, os.IsNotExist(statErr), "unsupported file should be removed")
}

func TestUploadCheckFile_SizeAccounting(t *testing.T) {
	dir := t.TempDir()
	// Use known-good JPEG
	src := filepath.Clean("../../pkg/fs/testdata/directory/example.jpg")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Skip("asset missing; skip")
	}
	f := filepath.Join(dir, "a.jpg")
	assert.NoError(t, os.WriteFile(f, data, 0o600))
	size := int64(len(data))
	// Set remaining limit to size+1 so it does not hit the removal branch (which triggers on <=0)
	rem, err := UploadCheckFile(f, false, size+1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), rem)
}
