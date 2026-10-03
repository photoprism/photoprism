package workers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

func TestShare_StartStoredError(t *testing.T) {
	// The remote answers the removal of an expired copy with an error whose text is longer than the column.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("removal failed\nshare › admin " + strings.Repeat("x", 900)))
	}))
	t.Cleanup(server.Close)

	a := entity.Service{AccName: "Share Error " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav", AccShare: true,
		SharePath: "/", ShareExpires: 60, AccTimeout: "low", RetryLimit: 3}
	require.NoError(t, entity.Db().Create(&a).Error)
	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileShare{}, "service_id = ?", a.ID).Error)
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
	})

	f := entity.NewFileShare(1000000, a.ID, "/expired.jpg")
	f.Status = entity.FileShareShared
	require.NoError(t, f.Create())
	expired := time.Now().Add(-time.Hour).Truncate(time.Second)
	require.NoError(t, entity.Db().Model(&entity.FileShare{}).Where("service_id = ?", a.ID).
		UpdateColumn("updated_at", expired).Error)

	require.NoError(t, NewShare(config.TestConfig()).Start())

	var stored entity.FileShare
	require.NoError(t, entity.Db().Where("service_id = ?", a.ID).First(&stored).Error)
	assert.Equal(t, 1, stored.Errors)
	assert.LessOrEqual(t, len(stored.Error), txt.ClipError)
	assert.True(t, utf8.ValidString(stored.Error))
	assert.NotContains(t, stored.Error, "\n")
	assert.NotContains(t, stored.Error, "›")
	assert.Contains(t, stored.Error, "500 Internal Server Error")

	// The saved row leaves the head of the expired queue until it is due again.
	assert.Equal(t, entity.FileShareShared, stored.Status)
	assert.True(t, stored.UpdatedAt.After(expired))
}

func TestShare_StartUploadError(t *testing.T) {
	conf := config.TestConfig()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options })
	conf.Options().OriginalsPath = t.TempDir()

	// The remote accepts folders but drops the connection on upload, so the error names the remote file.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusCreated)
			return
		}
		if conn, _, err := w.(http.Hijacker).Hijack(); err == nil {
			_ = conn.Close()
		}
	}))
	t.Cleanup(server.Close)

	a := entity.Service{AccName: "Share Upload Error " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav", AccShare: true,
		SharePath: "/", AccTimeout: "low", RetryLimit: 3}
	require.NoError(t, entity.Db().Create(&a).Error)

	photo := entity.NewPhoto(false)
	require.NoError(t, photo.Save())
	file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: "upload.jpg",
		FileType: fs.ImageJpeg.String(), MediaType: media.Image.String(), FileHash: fmt.Sprintf("%040d", 990)}
	require.NoError(t, file.Create())
	require.NoError(t, os.WriteFile(filepath.Join(conf.OriginalsPath(), file.FileName), []byte("upload"), fs.ModeFile))

	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileShare{}, "service_id = ?", a.ID).Error)
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(photo)
	})

	remoteName := "/" + strings.Repeat("\xff", 250) + ".jpg"
	require.NotNil(t, entity.FirstOrCreateFileShare(entity.NewFileShare(file.ID, a.ID, remoteName)))

	require.NoError(t, NewShare(conf).Start())

	var stored entity.FileShare
	require.NoError(t, entity.Db().Where("service_id = ?", a.ID).First(&stored).Error)
	assert.Equal(t, 1, stored.Errors)
	assert.LessOrEqual(t, len(stored.Error), txt.ClipError)
	assert.True(t, strings.HasPrefix(stored.Error, "webdav: failed to upload "))
}
