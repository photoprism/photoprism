package workers

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service/webdav"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestStallBudget(t *testing.T) {
	// Transfers that did not stall in between do not reset the count.
	var s stallBudget
	assert.False(t, s.Add(true))
	assert.False(t, s.Add(false))
	assert.False(t, s.Add(true))
	assert.False(t, s.Add(false))
	assert.True(t, s.Add(true))
	assert.True(t, s.Add(false))
}

// stallServer returns a server that stalls every transfer of a name that contains "stall", unless it ends with
// ".jpg", answers a download of such a file with a JPEG image, and counts the transfers per name; blocked requests
// are released when the test ends.
func stallServer(t *testing.T) (*httptest.Server, func(name string) int) {
	t.Helper()

	t.Cleanup(webdav.SetTransferStallTimeout(100 * time.Millisecond))

	var mu sync.Mutex
	requests := make(map[string]int)
	release := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPut {
			w.WriteHeader(http.StatusCreated)
			return
		}

		mu.Lock()
		requests[r.URL.Path]++
		mu.Unlock()

		if r.Method == http.MethodPut {
			_, _ = io.Copy(io.Discard, r.Body)
		}

		if strings.Contains(r.URL.Path, "stall") && !strings.HasSuffix(r.URL.Path, ".jpg") {
			select {
			case <-release:
			case <-r.Context().Done():
			}
			return
		}

		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusCreated)
			return
		}

		img := image.NewRGBA(image.Rect(0, 0, 8, 8))
		_ = jpeg.Encode(w, img, nil)
	}))
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(release) })

	return server, func(name string) int {
		mu.Lock()
		defer mu.Unlock()
		return requests[name]
	}
}

func TestSync_downloadStalls(t *testing.T) {
	conf := config.TestConfig()
	worker := NewSync(conf)
	server, requests := stallServer(t)

	// newAccount creates a download account on the test server with the specified remote files.
	newAccount := func(t *testing.T, names ...string) entity.Service {
		a := entity.Service{AccName: "Sync Stall " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav",
			AccSync: true, SyncDownload: true, SyncPath: "/", AccTimeout: "low", RetryLimit: -1}
		require.NoError(t, entity.Db().Create(&a).Error)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error)
			assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
			_ = os.RemoveAll(filepath.Join(worker.downloadPath(), fmt.Sprintf("%d", a.ID)))
		})
		for _, name := range names {
			require.NoError(t, entity.NewFileSync(a.ID, name).Create())
		}
		return a
	}

	t.Run("Budget", func(t *testing.T) {
		// The run stops at the third stalled transfer, also when others completed in between.
		names := []string{"/a-stall.txt", "/b-stall.txt", "/c-stall.txt", "/d-stall.txt", "/e-stall.txt"}
		a := newAccount(t, names...)
		complete, err := worker.download(a)
		require.NoError(t, err)
		assert.False(t, complete)

		// Files of different groups are read in no particular order.
		var total int
		for _, name := range names {
			total += requests(name)
		}
		assert.Equal(t, transferStallLimit, total)
	})
	t.Run("ImportsDownloaded", func(t *testing.T) {
		// Files of a group that downloaded before the run stopped are still imported. The files of one group are
		// read in name order, so the image comes before the stalled files and one is left.
		var mu sync.Mutex
		var imported []string
		prev := syncImportWorker
		syncImportWorker = func(jobs <-chan photoprism.ImportJob) {
			for job := range jobs {
				mu.Lock()
				imported = append(imported, filepath.Base(job.FileName))
				mu.Unlock()
			}
		}
		t.Cleanup(func() { syncImportWorker = prev })

		names := []string{"/grp-stall.avif", "/grp-stall.jpg", "/grp-stall.png", "/grp-stall.tif", "/grp-stall.webp"}
		a := newAccount(t, names...)
		_, err := worker.download(a)
		require.NoError(t, err)
		assert.Equal(t, 1, requests("/grp-stall.tif"))
		assert.Equal(t, 0, requests("/grp-stall.webp"))

		// The import worker reads its jobs asynchronously.
		require.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return len(imported) == 1
		}, 5*time.Second, 20*time.Millisecond)
		assert.Equal(t, []string{"grp-stall.jpg"}, imported)
	})
}

// newStallUploadAccount creates an upload or share account on the server with local files of the specified
// names, and records the fixture files as handled, so only those files are transferred.
func newStallUploadAccount(t *testing.T, conf *config.Config, server *httptest.Server, share bool, names ...string) (entity.Service, []*entity.File) {
	t.Helper()

	a := entity.Service{AccName: "Upload Stall " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav",
		AccSync: !share, SyncUpload: !share, AccShare: share, SyncPath: "/", SharePath: "/", AccTimeout: "low", RetryLimit: 3}
	require.NoError(t, entity.Db().Create(&a).Error)

	photo := entity.NewPhoto(false)
	require.NoError(t, photo.Save())

	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.FileShare{}, "service_id = ?", a.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&photo)
	})

	if !share {
		var fixtures entity.Files
		require.NoError(t, entity.Db().Find(&fixtures).Error)
		for _, file := range fixtures {
			record := entity.NewFileSync(a.ID, fmt.Sprintf("fixture/%d", file.ID))
			record.FileID, record.Status = file.ID, entity.FileSyncIgnore
			require.NoError(t, record.Create())
		}
	}

	var files []*entity.File

	for i, name := range names {
		file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: name,
			FileType: fs.ImageJpeg.String(), MediaType: media.Image.String(), FileHash: fmt.Sprintf("%040d", 1200+i)}
		require.NoError(t, file.Create())
		require.NoError(t, os.WriteFile(filepath.Join(conf.OriginalsPath(), name), []byte(name), fs.ModeFile))
		files = append(files, file)

		if share {
			require.NotNil(t, entity.FirstOrCreateFileShare(entity.NewFileShare(file.ID, a.ID, "/"+name)))
		}
	}

	return a, files
}

func TestSync_uploadStalls(t *testing.T) {
	conf := config.TestConfig()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options })
	conf.Options().OriginalsPath = t.TempDir()

	worker := NewSync(conf)
	server, requests := stallServer(t)
	names := []string{"a-stall.png", "b-stall.png", "c-stall.png", "d-stall.png"}
	a, _ := newStallUploadAccount(t, conf, server, false, names...)

	// The run stops at the third stalled upload.
	complete, err := worker.upload(a)
	require.NoError(t, err)
	assert.False(t, complete)
	assert.Equal(t, 1, requests("/c-stall.png"))
	assert.Equal(t, 0, requests("/d-stall.png"))

	// The next run uploads the other files first, and completes once only stalled files are left.
	complete, err = worker.upload(a)
	require.NoError(t, err)
	assert.False(t, complete)
	assert.Equal(t, 1, requests("/a-stall.png"))
	assert.Equal(t, 1, requests("/d-stall.png"))

	complete, err = worker.upload(a)
	require.NoError(t, err)
	assert.True(t, complete)

	// A refresh at the start of the next sync cycle selects them again.
	_, _ = worker.refresh(a)
	var stalled int64
	require.NoError(t, entity.Db().Model(&entity.FileSync{}).Where("service_id = ? AND remote_name LIKE ?",
		a.ID, entity.FileSyncStalledPrefix+"%").Count(&stalled).Error)
	assert.Zero(t, stalled)

	_, err = worker.upload(a)
	require.NoError(t, err)
	assert.Equal(t, 2, requests("/a-stall.png"))
}

func TestShare_StartStalls(t *testing.T) {
	conf := config.TestConfig()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options })
	conf.Options().OriginalsPath = t.TempDir()

	server, requests := stallServer(t)
	names := []string{"a-stall.png", "b-stall.png", "c-stall.png", "d-stall.png"}
	a, files := newStallUploadAccount(t, conf, server, true, names...)
	require.NoError(t, a.Update("RetryLimit", -1))

	// The run stops at the third stalled upload, and the remaining file stays queued.
	require.NoError(t, NewShare(conf).Start())
	assert.Equal(t, 1, requests("/c-stall.png"))
	assert.Equal(t, 0, requests("/d-stall.png"))

	var last entity.FileShare
	require.NoError(t, entity.Db().Where("service_id = ? AND file_id = ?", a.ID, files[3].ID).First(&last).Error)
	assert.Equal(t, entity.FileShareNew, last.Status)
	assert.Equal(t, 0, last.Errors)

	// The next run tries the file that has not failed first.
	require.NoError(t, NewShare(conf).Start())
	assert.Equal(t, 1, requests("/d-stall.png"))
	assert.Equal(t, 1, requests("/c-stall.png"))
}
