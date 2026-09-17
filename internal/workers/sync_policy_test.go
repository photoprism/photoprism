package workers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/webdav"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestSyncPathPolicy checks queued uploads/downloads and continued progress after exclusions.
func TestSyncPathPolicy(t *testing.T) {
	conf := config.TestConfig()
	options := *conf.Options()

	t.Cleanup(func() { *conf.Options() = options })

	conf.Options().OriginalsPath = t.TempDir()
	remote := t.TempDir()

	var transfers atomic.Int64
	var deletes atomic.Int64

	handler := &webdav.Handler{FileSystem: webdav.Dir(remote), LockSystem: webdav.NewMemLS()}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			deletes.Add(1)
		}
		if r.Method == "PUT" || r.Method == "GET" {
			transfers.Add(1)
		}
		handler.ServeHTTP(w, r)
	}))

	t.Cleanup(server.Close)

	a := entity.Service{AccName: "Path Policy", AccURL: server.URL + "/", AccType: "webdav", AccShare: true, SyncPath: "/", SyncRaw: true, SyncFilenames: true, AccTimeout: "low", RetryLimit: 3}
	require.NoError(t, entity.Db().Create(&a).Error)

	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.FileShare{}, "service_id = ?", a.ID)
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

	var owned []*entity.File

	names := []string{".config/source.jpg", ".photoprism/source.jpg"}
	reserved := fs.ReservedPathNames()

	for i := 0; i < 249; i++ {
		names = append(names, fmt.Sprintf("0blocked/%s/batch-%03d.jpg", reserved[i%len(reserved)], i))
	}

	names = append(names, "alias.jpg", "linked.jpg", "ordinary.jpg")
	external := filepath.Join(t.TempDir(), "external.jpg")
	require.NoError(t, os.WriteFile(external, []byte("linked-control"), fs.ModeFile))

	for i, name := range names {
		file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: name, FileType: "jpg", MediaType: entity.MediaImage, FileHash: fmt.Sprintf("%040d", i+700)}
		require.NoError(t, file.Create())
		owned = append(owned, file)
		filename := filepath.Join(conf.OriginalsPath(), name)
		require.NoError(t, fs.MkdirAll(filepath.Dir(filename)))
		switch name {
		case "alias.jpg":
			require.NoError(t, os.Symlink(filepath.Join(conf.OriginalsPath(), ".config/source.jpg"), filename))
		case "linked.jpg":
			require.NoError(t, os.Symlink(external, filename))
		default:
			require.NoError(t, os.WriteFile(filename, []byte(name), fs.ModeFile))
		}
	}

	duplicate := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootSidecar, FileName: ".config/source.jpg", FileType: "jpg", MediaType: entity.MediaImage, FileHash: "d943657924bbc9a043758915db118f9739d2fe11"}
	require.NoError(t, duplicate.Create())
	worker := NewSync(conf)
	complete, err := worker.upload(a)
	require.NoError(t, err)
	assert.False(t, complete)
	pending, err := query.AccountUploads(a, 250)
	require.NoError(t, err)
	assert.Len(t, pending, 5)
	assert.Zero(t, transfers.Load())
	complete, err = worker.upload(a)
	require.NoError(t, err)
	assert.False(t, complete)
	pending, err = query.AccountUploads(a, 250)
	require.NoError(t, err)
	assert.Empty(t, pending)
	complete, err = worker.upload(a)
	require.NoError(t, err)
	assert.True(t, complete)
	assert.EqualValues(t, 3, transfers.Load())
	assert.FileExists(t, filepath.Join(remote, "alias.jpg"))
	linked, err := os.ReadFile(filepath.Join(remote, "linked.jpg")) //nolint:gosec // Test reads its temporary transfer output.
	require.NoError(t, err)
	assert.Equal(t, "linked-control", string(linked))
	assert.FileExists(t, filepath.Join(remote, "ordinary.jpg"))
	assert.NoDirExists(t, filepath.Join(remote, ".config"))

	for _, file := range owned[:2] {
		var row entity.FileSync
		require.NoError(t, entity.Db().Where("service_id = ? AND file_id = ?", a.ID, file.ID).First(&row).Error)
		assert.Equal(t, entity.FileSyncIgnore, row.Status)
		assert.Empty(t, row.Error)
		assert.Zero(t, row.Errors)
	}

	for _, name := range append(fs.ReservedPathNames(), ".ssh/key.jpg", ".photoprism/metadata.json", ".hidden/ordinary.jpg") {
		require.NoError(t, entity.NewFileSync(a.ID, name).Create())
	}

	before := transfers.Load()
	complete, err = worker.download(a)

	require.NoError(t, err)
	assert.False(t, complete)
	assert.Equal(t, before, transfers.Load())

	rows, err := query.FileSyncs(a.ID, entity.FileSyncNew, 1000)

	require.NoError(t, err)
	assert.Empty(t, rows)

	complete, err = worker.download(a)

	require.NoError(t, err)
	assert.True(t, complete)

	a.SyncPath = ".config"
	complete, err = worker.refresh(a)

	require.NoError(t, err)
	assert.True(t, complete)

	complete, err = worker.upload(a)

	require.NoError(t, err)
	assert.True(t, complete)

	complete, err = worker.download(a)

	require.NoError(t, err)
	assert.True(t, complete)
	assert.Equal(t, before, transfers.Load())

	a.SyncPath = "/"

	require.NoError(t, os.Symlink(filepath.Join(conf.OriginalsPath(), ".config"), filepath.Join(conf.OriginalsPath(), "download-alias")))
	require.NoError(t, fs.MkdirAll(filepath.Join(remote, "download-alias")))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "download-alias/remote.txt"), []byte("excluded-control"), fs.ModeFile))
	require.NoError(t, os.WriteFile(filepath.Join(remote, "fresh.txt"), []byte("download-control"), fs.ModeFile))
	require.NoError(t, fs.MkdirAll(filepath.Join(remote, ".ssh")))
	require.NoError(t, os.WriteFile(filepath.Join(remote, ".ssh/remote.txt"), []byte("excluded-control"), fs.ModeFile))

	complete, err = worker.refresh(a)

	require.NoError(t, err)
	assert.True(t, complete)

	queued, err := query.FileSyncs(a.ID, entity.FileSyncNew, 1000)

	require.NoError(t, err)
	require.Len(t, queued, 2)

	beforeDownload := transfers.Load()
	_, err = worker.download(a)

	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(conf.OriginalsPath(), "fresh.txt")) //nolint:gosec // Test reads a controlled fixture or generated output.

	require.NoError(t, err)
	assert.Equal(t, "download-control", string(data))
	assert.Equal(t, beforeDownload+2, transfers.Load())
	assert.FileExists(t, filepath.Join(conf.OriginalsPath(), ".config/remote.txt"))
	assert.NoFileExists(t, filepath.Join(conf.OriginalsPath(), ".ssh/remote.txt"))
	require.NoError(t, entity.NewFileShare(owned[len(owned)-3].ID, a.ID, "symlink-shared.jpg").Create())
	require.NoError(t, entity.NewFileShare(owned[len(owned)-2].ID, a.ID, "linked-shared.jpg").Create())
	require.NoError(t, entity.NewFileShare(owned[0].ID, a.ID, "ordinary-alias.jpg").Create())
	require.NoError(t, entity.NewFileShare(owned[len(owned)-1].ID, a.ID, ".ssh/ordinary.jpg").Create())
	require.NoError(t, entity.NewFileShare(owned[len(owned)-1].ID, a.ID, "shared.jpg").Create())
	broken := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: "broken.jpg", FileType: "jpg", MediaType: entity.MediaImage, FileHash: "920de6c3667435aab021245438a879e53374c43d"}
	require.NoError(t, broken.Create())
	require.NoError(t, os.Symlink(filepath.Join(conf.OriginalsPath(), "absent.jpg"), filepath.Join(conf.OriginalsPath(), broken.FileName)))
	brokenShare := entity.NewFileShare(broken.ID, a.ID, "broken-share.jpg")
	brokenShare.Errors = a.RetryLimit
	require.NoError(t, brokenShare.Create())
	require.NoError(t, NewShare(conf).Start())
	storedBroken := entity.FileShare{}
	require.NoError(t, entity.Db().Where("service_id = ? AND file_id = ?", a.ID, broken.ID).First(&storedBroken).Error)
	assert.Equal(t, entity.FileShareError, storedBroken.Status)
	assert.Equal(t, a.RetryLimit+1, storedBroken.Errors)

	assert.NoFileExists(t, filepath.Join(remote, "ordinary-alias.jpg"))
	assert.NoFileExists(t, filepath.Join(remote, ".ssh/ordinary.jpg"))
	assert.FileExists(t, filepath.Join(remote, "shared.jpg"))

	var ignored []entity.FileShare

	require.NoError(t, entity.Db().Where("service_id = ? AND status = ?", a.ID, entity.FileShareIgnore).Find(&ignored).Error)
	assert.Len(t, ignored, 2)
	assert.FileExists(t, filepath.Join(remote, "symlink-shared.jpg"))
	assert.FileExists(t, filepath.Join(remote, "linked-shared.jpg"))

	for _, record := range ignored {
		assert.Empty(t, record.Error)
		assert.Zero(t, record.Errors)
	}

	// Expired protected copies require manual removal, while eligible copies are deleted.
	require.NoError(t, entity.Db().Model(&a).Update("ShareExpires", 3600).Error)
	for _, name := range []string{".ssh/expired.jpg", "expired.jpg"} {
		require.NoError(t, os.WriteFile(filepath.Join(remote, name), []byte("expiry-control"), fs.ModeFile))
		row := entity.NewFileShare(owned[len(owned)-1].ID, a.ID, name)
		row.Status = entity.FileShareShared
		require.NoError(t, row.Create())
		require.NoError(t, entity.Db().Model(row).UpdateColumn("updated_at", time.Now().UTC().Add(-48*time.Hour)).Error)
	}
	require.NoError(t, NewShare(conf).Start())
	assert.EqualValues(t, 1, deletes.Load())
	assert.FileExists(t, filepath.Join(remote, ".ssh/expired.jpg"))
	assert.NoFileExists(t, filepath.Join(remote, "expired.jpg"))
	var retained entity.FileShare
	require.NoError(t, entity.Db().Where("service_id = ? AND remote_name = ?", a.ID, ".ssh/expired.jpg").First(&retained).Error)
	assert.Equal(t, entity.FileShareError, retained.Status)
	assert.Contains(t, retained.Error, "remote copy retained")
	assert.Equal(t, 1, retained.Errors)
	require.NoError(t, NewShare(conf).Start())
	assert.EqualValues(t, 1, deletes.Load())
}
