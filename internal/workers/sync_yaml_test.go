package workers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/webdav"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// TestSyncYaml checks that a refused YAML upload disables YAML sync while other files keep syncing.
func TestSyncYaml(t *testing.T) {
	conf := config.TestConfig()
	options := *conf.Options()

	t.Cleanup(func() { *conf.Options() = options })

	conf.Options().OriginalsPath = t.TempDir()
	remote := t.TempDir()

	var yamlPuts, yamlGets atomic.Int64
	var refuseAll, failYaml atomic.Bool

	handler := &webdav.Handler{FileSystem: webdav.Dir(remote), LockSystem: webdav.NewMemLS()}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, fs.ExtYml) && r.Method == http.MethodPut {
			yamlPuts.Add(1)
		}
		if r.Method == http.MethodPut && failYaml.Load() && strings.HasSuffix(r.URL.Path, fs.ExtYml) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if r.Method == http.MethodPut && refuseAll.Load() {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if strings.HasSuffix(r.URL.Path, fs.ExtYml) {
			switch r.Method {
			case http.MethodPut:
				w.WriteHeader(http.StatusForbidden)
				return
			case http.MethodGet:
				yamlGets.Add(1)
			}
		}
		handler.ServeHTTP(w, r)
	}))

	t.Cleanup(server.Close)

	a := entity.Service{AccName: "Sync YAML", AccURL: server.URL + "/", AccType: "webdav", AccSync: true, SyncPath: "/", SyncFilenames: true, SyncYaml: true, AccTimeout: "low", RetryLimit: 3}
	require.NoError(t, entity.Db().Create(&a).Error)

	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID)
		entity.UnscopedDb().Unscoped().Delete(&a)
	})

	// Mark existing fixture files as already considered for this test-owned service.
	var fixtures entity.Files

	require.NoError(t, entity.Db().Find(&fixtures).Error)

	for _, file := range fixtures {
		record := entity.NewFileSync(a.ID, fmt.Sprintf("fixture/%d", file.ID))
		record.FileID = file.ID
		record.Status = entity.FileSyncIgnore
		require.NoError(t, record.Create())
	}

	photo := entity.NewPhoto(false)

	require.NoError(t, photo.Save())

	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(photo)
	})

	for i, name := range []string{"a.jpg", "a.yml", "b.yml", "c.jpg"} {
		fileType, mediaType := fs.ImageJpeg, media.Image

		if fs.FileType(name) == fs.SidecarYaml {
			fileType, mediaType = fs.SidecarYaml, media.Sidecar
		}

		file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: name,
			FileType: fileType.String(), MediaType: mediaType.String(), FileHash: fmt.Sprintf("%040d", i+900)}
		require.NoError(t, file.Create())
		require.NoError(t, os.WriteFile(filepath.Join(conf.OriginalsPath(), name), []byte(name), fs.ModeFile))
	}

	stored := func(t *testing.T) entity.Service {
		var m entity.Service
		require.NoError(t, entity.Db().First(&m, a.ID).Error)
		return m
	}

	worker := NewSync(conf)

	t.Run("RemoteRefusesAll", func(t *testing.T) {
		refuseAll.Store(true)
		defer refuseAll.Store(false)
		yamlPuts.Store(0)
		_, err := worker.upload(a)
		require.NoError(t, err)
		assert.EqualValues(t, 1, yamlPuts.Load(), "a YAML file must not be sent after one was refused")
		assert.NoFileExists(t, filepath.Join(remote, "a.jpg"))
		assert.True(t, stored(t).SyncYaml, "YAML sync must stay on when other files are refused as well")
	})
	t.Run("UploadDisablesYaml", func(t *testing.T) {
		yamlPuts.Store(0)
		complete, err := worker.upload(a)
		require.NoError(t, err)
		assert.False(t, complete)
		assert.EqualValues(t, 1, yamlPuts.Load(), "a YAML file must not be sent after one was refused")
		assert.FileExists(t, filepath.Join(remote, "a.jpg"))
		assert.FileExists(t, filepath.Join(remote, "c.jpg"))
		a = stored(t)
		assert.False(t, a.SyncYaml)
	})
	t.Run("UploadCompletes", func(t *testing.T) {
		pending, err := query.AccountUploads(a, 250)
		require.NoError(t, err)
		assert.Empty(t, pending)
		complete, err := worker.upload(a)
		require.NoError(t, err)
		assert.True(t, complete)
		assert.EqualValues(t, 1, yamlPuts.Load())
	})
	t.Run("RemoteFailsYaml", func(t *testing.T) {
		failYaml.Store(true)
		defer failYaml.Store(false)
		require.NoError(t, a.Update("SyncYaml", true))
		a.SyncYaml = true
		defer func() {
			a.SyncYaml = false
			require.NoError(t, a.Update("SyncYaml", false))
		}()
		yamlPuts.Store(0)
		_, err := worker.upload(a)
		require.NoError(t, err)
		assert.EqualValues(t, 2, yamlPuts.Load(), "YAML files must be retried after other errors")
		assert.True(t, stored(t).SyncYaml)
	})

	require.NoError(t, os.WriteFile(filepath.Join(remote, "remote.yml"), []byte("remote-yaml"), fs.ModeFile))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "remote.txt"), []byte("remote-text"), fs.ModeFile))

	status := func(t *testing.T, name string) string {
		var row entity.FileSync
		require.NoError(t, entity.Db().Where("service_id = ? AND remote_name = ?", a.ID, name).First(&row).Error)
		return row.Status
	}

	t.Run("RefreshIgnoresYaml", func(t *testing.T) {
		complete, err := worker.refresh(a)
		require.NoError(t, err)
		assert.True(t, complete)
		assert.Equal(t, entity.FileSyncIgnore, status(t, "/remote.yml"))
		assert.Equal(t, entity.FileSyncNew, status(t, "/remote.txt"))
	})
	t.Run("RefreshRequeuesYaml", func(t *testing.T) {
		a.SyncYaml = true
		complete, err := worker.refresh(a)
		require.NoError(t, err)
		assert.True(t, complete)
		assert.Equal(t, entity.FileSyncNew, status(t, "/remote.yml"))
	})
	t.Run("DownloadSkipsYaml", func(t *testing.T) {
		a.SyncYaml = false
		_, err := worker.download(a)
		require.NoError(t, err)
		assert.Equal(t, entity.FileSyncIgnore, status(t, "/remote.yml"))
		assert.Equal(t, entity.FileSyncDownloaded, status(t, "/remote.txt"))
		assert.Zero(t, yamlGets.Load())
		assert.NoFileExists(t, filepath.Join(conf.OriginalsPath(), "remote.yml"))
		assert.FileExists(t, filepath.Join(conf.OriginalsPath(), "remote.txt"))
	})
}

// TestShareYaml checks that manual uploads keep YAML sidecar files queued while the account does not sync them.
func TestShareYaml(t *testing.T) {
	conf := config.TestConfig()
	options := *conf.Options()

	t.Cleanup(func() { *conf.Options() = options })

	conf.Options().OriginalsPath = t.TempDir()
	remote := t.TempDir()

	var yamlPuts atomic.Int64
	var refuseAll, acceptYaml atomic.Bool

	handler := &webdav.Handler{FileSystem: webdav.Dir(remote), LockSystem: webdav.NewMemLS()}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		yamlPut := r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, fs.ExtYml)
		if yamlPut {
			yamlPuts.Add(1)
		}
		if r.Method == http.MethodPut && refuseAll.Load() || yamlPut && !acceptYaml.Load() {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		handler.ServeHTTP(w, r)
	}))

	t.Cleanup(server.Close)

	a := entity.Service{AccName: "Share YAML", AccURL: server.URL + "/", AccType: "webdav", AccShare: true, SharePath: "/", SyncYaml: true, AccTimeout: "low", RetryLimit: 3}
	require.NoError(t, entity.Db().Create(&a).Error)

	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.FileShare{}, "service_id = ?", a.ID)
		entity.UnscopedDb().Unscoped().Delete(&a)
	})

	photo := entity.NewPhoto(false)

	require.NoError(t, photo.Save())

	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(photo)
	})

	files := make(map[string]*entity.File)

	for i, name := range []string{"a.jpg", "a.yml", "b.yml", "c.jpg", "d.yml", "e.jpg"} {
		fileType, mediaType := fs.ImageJpeg, media.Image

		if fs.FileType(name) == fs.SidecarYaml {
			fileType, mediaType = fs.SidecarYaml, media.Sidecar
		}

		file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: name,
			FileType: fileType.String(), MediaType: mediaType.String(), FileHash: fmt.Sprintf("%040d", i+950)}
		require.NoError(t, file.Create())
		require.NoError(t, os.WriteFile(filepath.Join(conf.OriginalsPath(), name), []byte(name), fs.ModeFile))
		files[name] = file
	}

	queue := func(t *testing.T, name string) {
		require.NotNil(t, entity.FirstOrCreateFileShare(entity.NewFileShare(files[name].ID, a.ID, name)))
	}

	share := func(t *testing.T, name string) entity.FileShare {
		var row entity.FileShare
		require.NoError(t, entity.Db().Where("service_id = ? AND file_id = ?", a.ID, files[name].ID).First(&row).Error)
		return row
	}

	syncYaml := func(t *testing.T) bool {
		var m entity.Service
		require.NoError(t, entity.Db().First(&m, a.ID).Error)
		return m.SyncYaml
	}

	worker := NewShare(conf)

	t.Run("RemoteRefusesAll", func(t *testing.T) {
		refuseAll.Store(true)
		defer refuseAll.Store(false)
		queue(t, "a.jpg")
		queue(t, "a.yml")
		require.NoError(t, worker.Start())
		assert.EqualValues(t, 1, yamlPuts.Load())
		assert.True(t, syncYaml(t), "YAML sync must stay on when other files are refused as well")
		assert.Equal(t, 1, share(t, "a.yml").Errors)
		assert.Equal(t, entity.FileShareNew, share(t, "a.yml").Status)
	})
	t.Run("RefusedYamlDisablesYaml", func(t *testing.T) {
		yamlPuts.Store(0)
		queue(t, "b.yml")
		queue(t, "c.jpg")
		require.NoError(t, worker.Start())
		assert.EqualValues(t, 1, yamlPuts.Load(), "a YAML file must not be sent after one was refused")
		assert.False(t, syncYaml(t))
		assert.Equal(t, entity.FileShareShared, share(t, "a.jpg").Status)
		assert.Equal(t, entity.FileShareShared, share(t, "c.jpg").Status)
		assert.Equal(t, entity.FileShareNew, share(t, "b.yml").Status)
		assert.Zero(t, share(t, "b.yml").Errors)
		assert.Equal(t, entity.FileShareNew, share(t, "a.yml").Status)
		assert.Equal(t, 1, share(t, "a.yml").Errors, "a refusal that disables YAML sync must not count as an error")
	})
	t.Run("YamlStaysQueued", func(t *testing.T) {
		yamlPuts.Store(0)
		require.NoError(t, worker.Start())
		assert.Zero(t, yamlPuts.Load())
		assert.Equal(t, entity.FileShareNew, share(t, "a.yml").Status)
		assert.Equal(t, entity.FileShareNew, share(t, "b.yml").Status)
	})
	t.Run("ReenabledYamlUploads", func(t *testing.T) {
		acceptYaml.Store(true)
		require.NoError(t, a.Update("SyncYaml", true))
		require.NoError(t, worker.Start())
		assert.Equal(t, entity.FileShareShared, share(t, "a.yml").Status)
		assert.Equal(t, entity.FileShareShared, share(t, "b.yml").Status)
		assert.FileExists(t, filepath.Join(remote, "b.yml"))
		assert.True(t, syncYaml(t))
	})
	t.Run("RetryLimit", func(t *testing.T) {
		refuseAll.Store(true)
		defer refuseAll.Store(false)
		queue(t, "d.yml")
		queue(t, "e.jpg")
		for i := 0; i <= a.RetryLimit; i++ {
			require.NoError(t, worker.Start())
		}
		assert.Equal(t, entity.FileShareError, share(t, "d.yml").Status)
		assert.Equal(t, a.RetryLimit+1, share(t, "d.yml").Errors)
		assert.True(t, syncYaml(t))
	})
}
