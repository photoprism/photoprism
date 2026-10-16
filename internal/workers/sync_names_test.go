package workers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/webdav"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// captureSyncLog replaces the package logger until the test ends and returns its hook.
func captureSyncLog(t *testing.T) *test.Hook {
	t.Helper()

	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	prev := log
	log = logger
	t.Cleanup(func() { log = prev })

	return hook
}

// countLogged returns the number of entries with the specified level that contain text.
func countLogged(hook *test.Hook, level logrus.Level, text string) (n int) {
	for _, e := range hook.AllEntries() {
		if e.Level == level && strings.Contains(e.Message, text) {
			n++
		}
	}

	return n
}

// remoteNameOfLen returns a path of exactly n bytes that starts with prefix and ends with ext, using folder
// and file names of at most 200 bytes.
func remoteNameOfLen(prefix string, n int, ext string) string {
	name := prefix

	for n-len(name)-len(ext) > 200 {
		name += strings.Repeat("x", 199) + "/"
	}

	return name + strings.Repeat("y", n-len(name)-len(ext)) + ext
}

// renamedFS serves a local folder, listing the files named in names under other names, so that the test
// server can list names that the local file system cannot store.
type renamedFS struct {
	webdav.Dir
	names map[string]string
}

// local returns the local path of a listed name.
func (f renamedFS) local(name string) string {
	for short, long := range f.names {
		name = strings.Replace(name, long, short, 1)
	}

	return name
}

// listed returns file info with the listed name of a local file.
func (f renamedFS) listed(fi os.FileInfo) os.FileInfo {
	if fi == nil {
		return nil
	} else if long, ok := f.names[fi.Name()]; ok {
		return renamedInfo{FileInfo: fi, name: long}
	}

	return fi
}

// OpenFile opens the local file of a listed name.
func (f renamedFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	file, err := f.Dir.OpenFile(ctx, f.local(name), flag, perm)

	if err != nil {
		return nil, err
	}

	return renamedFile{File: file, fs: f}, nil
}

// Stat returns the file info of a listed name.
func (f renamedFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	fi, err := f.Dir.Stat(ctx, f.local(name))
	return f.listed(fi), err
}

// renamedFile lists the entries of a folder under their listed names.
type renamedFile struct {
	webdav.File
	fs renamedFS
}

// Readdir returns the entries of a folder under their listed names.
func (f renamedFile) Readdir(count int) ([]os.FileInfo, error) {
	list, err := f.File.Readdir(count)

	for i := range list {
		list[i] = f.fs.listed(list[i])
	}

	return list, err
}

// Stat returns the file info under the listed name.
func (f renamedFile) Stat() (os.FileInfo, error) {
	fi, err := f.File.Stat()
	return f.fs.listed(fi), err
}

// renamedInfo is file info with a listed name.
type renamedInfo struct {
	os.FileInfo
	name string
}

// Name returns the listed name.
func (i renamedInfo) Name() string { return i.name }

func TestSync_longRemoteNames(t *testing.T) {
	conf := config.TestConfig()
	worker := NewSync(conf)
	remote := t.TempDir()

	// The remote lists two files under names of 255 and 256 bytes.
	atLimit, overLimit := strings.Repeat("n", 251)+".txt", strings.Repeat("o", 252)+".txt"
	renamed := renamedFS{Dir: webdav.Dir(remote), names: map[string]string{"at-limit.txt": atLimit, "over-limit.txt": overLimit}}

	var puts, gets atomic.Int64

	handler := &webdav.Handler{FileSystem: renamed, LockSystem: webdav.NewMemLS()}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			puts.Add(1)
		case http.MethodGet:
			gets.Add(1)
		}
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)

	// newAccount creates an account on the test server and deletes it with its records when the test ends.
	newAccount := func(t *testing.T, update func(a *entity.Service)) entity.Service {
		a := entity.Service{AccName: "Sync Names " + rnd.Base36(8), AccURL: server.URL + "/", AccType: "webdav", AccSync: true,
			SyncPath: "/", AccTimeout: "low", RetryLimit: -1}
		update(&a)
		require.NoError(t, entity.Db().Create(&a).Error)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error)
			assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
			_ = os.RemoveAll(filepath.Join(worker.downloadPath(), strconv.FormatUint(uint64(a.ID), 10)))
		})
		return a
	}

	t.Run("Download", func(t *testing.T) {
		a := newAccount(t, func(a *entity.Service) { a.SyncDownload = true })

		// Names up to the limits are queued, a longer one, and one with a file name over 255 bytes, are not.
		queued := []string{remoteNameOfLen("/names/mid/", 300, ".txt"), remoteNameOfLen("/names/limit/", entity.RemoteNameMaxLen, ".txt"), "/names/" + atLimit}
		skipped := []string{remoteNameOfLen("/names/over/", entity.RemoteNameMaxLen+1, ".txt"), "/names/" + overLimit}

		for short := range renamed.names {
			require.NoError(t, fs.MkdirAll(filepath.Join(remote, "names")))
			require.NoError(t, os.WriteFile(filepath.Join(remote, "names", short), []byte(short), fs.ModeFile))
		}

		for _, name := range []string{queued[0], queued[1], skipped[0]} {
			fileName := filepath.Join(remote, filepath.FromSlash(name))
			require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))
			require.NoError(t, os.WriteFile(fileName, []byte(name[len(name)-12:]), fs.ModeFile))
		}

		hook := captureSyncLog(t)

		for run := 1; run <= 2; run++ {
			complete, err := worker.refresh(a)
			require.NoError(t, err)
			assert.True(t, complete)

			var names []string
			require.NoError(t, entity.Db().Model(&entity.FileSync{}).Where("service_id = ?", a.ID).Pluck("remote_name", &names).Error)
			assert.ElementsMatch(t, queued, names)

			assert.Equal(t, run, countLogged(hook, logrus.WarnLevel, ""))
			assert.Equal(t, run, countLogged(hook, logrus.WarnLevel, "skipped 2 remote files in"), hook.LastEntry().Message)
			assert.Equal(t, run, countLogged(hook, logrus.WarnLevel, a.AccName))
			assert.Equal(t, run, countLogged(hook, logrus.WarnLevel, ".txt"))
			assert.Zero(t, countLogged(hook, logrus.ErrorLevel, ""))
		}

		_, err := worker.download(a)
		require.NoError(t, err)

		for _, name := range queued {
			var f entity.FileSync
			require.NoError(t, entity.Db().Where("service_id = ? AND remote_name = ?", a.ID, name).First(&f).Error)
			assert.Equal(t, entity.FileSyncDownloaded, f.Status)
			assert.FileExists(t, filepath.Join(worker.downloadPath(), strconv.FormatUint(uint64(a.ID), 10), filepath.FromSlash(name)))
		}
	})
	t.Run("QueuedLongName", func(t *testing.T) {
		// A queued file whose name is too long for local file systems is ignored without a request.
		a := newAccount(t, func(a *entity.Service) { a.SyncDownload = true })
		name := "/queued/" + overLimit
		require.NoError(t, entity.NewFileSync(a.ID, name).Create())
		before := gets.Load()

		hook := captureSyncLog(t)
		_, err := worker.download(a)
		require.NoError(t, err)

		var f entity.FileSync
		require.NoError(t, entity.Db().Where("service_id = ? AND remote_name = ?", a.ID, name).First(&f).Error)
		assert.Equal(t, entity.FileSyncIgnore, f.Status)
		assert.Equal(t, before, gets.Load())
		assert.Zero(t, countLogged(hook, logrus.ErrorLevel, ""))
	})
	// newUploadPhoto records the fixture files as handled for an account, so only the files of the test are
	// uploaded, and returns a new photo for them that is deleted with its files when the test ends.
	newUploadPhoto := func(t *testing.T, a entity.Service) entity.Photo {
		var fixtures entity.Files
		require.NoError(t, entity.Db().Find(&fixtures).Error)
		for _, file := range fixtures {
			record := entity.NewFileSync(a.ID, fmt.Sprintf("fixture/%d", file.ID))
			record.FileID, record.Status = file.ID, entity.FileSyncIgnore
			require.NoError(t, record.Create())
		}

		photo := entity.NewPhoto(false)
		require.NoError(t, photo.Save())
		t.Cleanup(func() {
			entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", photo.ID)
			entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
			entity.UnscopedDb().Unscoped().Delete(&photo)
		})

		return photo
	}

	t.Run("UploadLongFileName", func(t *testing.T) {
		// The remote decides whether it accepts a long file name, so the file is tried rather than ignored.
		a := newAccount(t, func(a *entity.Service) { a.SyncUpload = true })
		photo := newUploadPhoto(t, a)
		file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals,
			FileName: "long/" + strings.Repeat("l", 256) + ".jpg", FileType: "jpg", MediaType: entity.MediaImage, FileHash: fmt.Sprintf("%040d", 790)}
		require.NoError(t, file.Create())

		hook := captureSyncLog(t)
		_, err := worker.upload(a)
		require.NoError(t, err)
		assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, "failed to upload 1 file to "+clean.Log(a.AccName)))

		// The failed upload is set aside until the next sync cycle.
		var m entity.FileSync
		require.NoError(t, entity.Db().Where("service_id = ? AND file_id = ?", a.ID, file.ID).First(&m).Error)
		assert.Equal(t, entity.FileSyncUploadName(file.ID), m.RemoteName)
		assert.Equal(t, entity.FileSyncFailed, m.Status)
	})
	t.Run("Upload", func(t *testing.T) {
		options := *conf.Options()
		t.Cleanup(func() { *conf.Options() = options })
		conf.Options().OriginalsPath = t.TempDir()

		syncPath := remoteNameOfLen("/", 300, "/upload")
		a := newAccount(t, func(a *entity.Service) {
			a.SyncUpload = true
			a.SyncPath = syncPath
		})

		photo := newUploadPhoto(t, a)

		// Below the sync path, the first two names fit the limit, the last two exceed it.
		room := entity.RemoteNameMaxLen - len(syncPath) - 1
		names := []string{"upload-mid.jpg", remoteNameOfLen("limit/", room, ".jpg"),
			remoteNameOfLen("over/", room+1, ".jpg"), remoteNameOfLen("long/", room+100, ".jpg")}
		var files []*entity.File

		for i, name := range names {
			file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: name,
				FileType: "jpg", MediaType: entity.MediaImage, FileHash: fmt.Sprintf("%040d", i+800)}
			require.NoError(t, file.Create())
			files = append(files, file)
			fileName := filepath.Join(conf.OriginalsPath(), filepath.FromSlash(name))
			require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))
			require.NoError(t, os.WriteFile(fileName, []byte(name[len(name)-10:]), fs.ModeFile))
		}

		require.Greater(t, len(path.Join(syncPath, names[0])), 255)
		require.Len(t, path.Join(syncPath, names[1]), entity.RemoteNameMaxLen)

		hook := captureSyncLog(t)

		complete, err := worker.upload(a)
		require.NoError(t, err)
		assert.False(t, complete)
		assert.EqualValues(t, 2, puts.Load())
		assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, "skipped upload of 2 files to"))

		for _, file := range files[:2] {
			var uploaded entity.FileSync
			require.NoError(t, entity.Db().Where("service_id = ? AND file_id = ?", a.ID, file.ID).First(&uploaded).Error)
			assert.Equal(t, path.Join(syncPath, file.FileName), uploaded.RemoteName)
			assert.Equal(t, entity.FileSyncUploaded, uploaded.Status)
			assert.FileExists(t, filepath.Join(remote, filepath.FromSlash(uploaded.RemoteName)))
		}

		for _, file := range files[2:] {
			var ignored entity.FileSync
			require.NoError(t, entity.Db().Where("service_id = ? AND file_id = ?", a.ID, file.ID).First(&ignored).Error)
			assert.Equal(t, entity.FileSyncIgnore, ignored.Status)
			assert.Equal(t, entity.FileSyncLongNamePrefix(syncPath)+strconv.FormatUint(uint64(file.ID), 10), ignored.RemoteName)
		}

		// The next run finds nothing left to upload.
		complete, err = worker.upload(a)
		require.NoError(t, err)
		assert.True(t, complete)
		assert.EqualValues(t, 2, puts.Load())
		assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, ""))
		assert.Zero(t, countLogged(hook, logrus.ErrorLevel, ""))

		// A shorter sync path makes room, so the next run selects the skipped files again.
		f, err := form.NewService(a)
		require.NoError(t, err)
		f.SyncPath = "/short"
		require.NoError(t, a.SaveForm(f))

		complete, err = worker.upload(a)
		require.NoError(t, err)
		assert.False(t, complete)
		assert.EqualValues(t, 4, puts.Load())

		for _, file := range files {
			var count int64
			require.NoError(t, entity.Db().Model(&entity.FileSync{}).Where("service_id = ? AND file_id = ? AND status = ?",
				a.ID, file.ID, entity.FileSyncUploaded).Count(&count).Error)
			assert.EqualValues(t, 1, count, file.FileName)
		}
	})
	t.Run("UploadUnportable", func(t *testing.T) {
		options := *conf.Options()
		t.Cleanup(func() { *conf.Options() = options })
		conf.Options().OriginalsPath = t.TempDir()

		a := newAccount(t, func(a *entity.Service) {
			a.SyncUpload = true
			a.SyncPath = "/@upload-names"
		})

		photo := newUploadPhoto(t, a)
		names := []string{"2019/portable.jpg", "@Family/__x.jpg", "2019/a:b.jpg", "dir*/x.jpg"}
		var files []*entity.File

		for i, name := range names {
			file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID, FileRoot: entity.RootOriginals, FileName: name,
				FileType: "jpg", MediaType: entity.MediaImage, FileHash: fmt.Sprintf("%040d", i+900)}
			require.NoError(t, file.Create())
			files = append(files, file)
			fileName := filepath.Join(conf.OriginalsPath(), filepath.FromSlash(name))
			require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))
			require.NoError(t, os.WriteFile(fileName, []byte(name), fs.ModeFile))
		}

		hook := captureSyncLog(t)
		before := puts.Load()

		_, err := worker.upload(a)
		require.NoError(t, err)
		assert.EqualValues(t, 2, puts.Load()-before)
		assert.FileExists(t, filepath.Join(remote, "@upload-names", "@Family", "__x.jpg"))
		assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, "skipped upload of 2 files to"))
		assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, "because of unsupported characters"))

		for _, file := range files[2:] {
			var ignored entity.FileSync
			require.NoError(t, entity.Db().Where("service_id = ? AND file_id = ?", a.ID, file.ID).First(&ignored).Error)
			assert.Equal(t, entity.FileSyncIgnore, ignored.Status)
			assert.Equal(t, entity.FileSyncUnportablePrefix+strconv.FormatUint(uint64(file.ID), 10), ignored.RemoteName)
		}

		// The next run uploads nothing and warns no more.
		_, err = worker.upload(a)
		require.NoError(t, err)
		assert.EqualValues(t, 2, puts.Load()-before)
		assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, "because of unsupported characters"))

		// A renamed file keeps its record and is uploaded under its new name.
		renamed := filepath.Join(conf.OriginalsPath(), "2019", "a-b.jpg")
		require.NoError(t, os.Rename(filepath.Join(conf.OriginalsPath(), "2019", "a:b.jpg"), renamed))
		require.NoError(t, files[2].Update("FileName", "2019/a-b.jpg"))

		_, err = worker.upload(a)
		require.NoError(t, err)
		assert.EqualValues(t, 3, puts.Load()-before)
		assert.FileExists(t, filepath.Join(remote, "@upload-names", "2019", "a-b.jpg"))
		assert.Equal(t, 1, countLogged(hook, logrus.WarnLevel, "because of unsupported characters"))
	})
}

func TestSync_ignoreUpload(t *testing.T) {
	worker := NewSync(config.TestConfig())
	a := entity.Service{AccName: "Sync Ignore " + rnd.Base36(8), AccType: "webdav"}
	require.NoError(t, entity.Db().Create(&a).Error)
	t.Cleanup(func() {
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.FileSync{}, "service_id = ?", a.ID).Error)
		assert.NoError(t, entity.UnscopedDb().Unscoped().Delete(&entity.Service{}, a.ID).Error)
	})

	hook := captureSyncLog(t)

	// Recording the same file twice keeps one record.
	worker.ignoreUpload(a, entity.FileSyncLongNamePrefix("/"), 4242)
	worker.ignoreUpload(a, entity.FileSyncLongNamePrefix("/"), 4242)

	var rows []entity.FileSync
	require.NoError(t, entity.Db().Where("service_id = ?", a.ID).Find(&rows).Error)
	require.Len(t, rows, 1)
	assert.Equal(t, uint(4242), rows[0].FileID)
	assert.Equal(t, entity.FileSyncIgnore, rows[0].Status)
	assert.Equal(t, entity.FileSyncLongNamePrefix("/")+"4242", rows[0].RemoteName)
	assert.Zero(t, countLogged(hook, logrus.ErrorLevel, ""))
}

func TestSkippedNames(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		hook := captureSyncLog(t)
		var s skippedNames
		s.Warn("sync: skipped %s in %s (%s)", "file", "Account")
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("Few", func(t *testing.T) {
		hook := captureSyncLog(t)
		var s skippedNames
		s.Add("/skipped-a.jpg")
		s.Add("/skipped-b.jpg")
		s.Warn("sync: skipped %s in %s (%s)", "file", "Account")
		require.Len(t, hook.AllEntries(), 1)
		msg := hook.LastEntry().Message
		assert.Contains(t, msg, "skipped 2 files in ")
		assert.Contains(t, msg, "skipped-b.jpg")
		assert.NotContains(t, msg, "more")
	})
	t.Run("Many", func(t *testing.T) {
		hook := captureSyncLog(t)
		var s skippedNames
		for i := 0; i < 12; i++ {
			s.Add(fmt.Sprintf("/skipped-%02d.jpg", i))
		}
		assert.Equal(t, 12, s.count)
		assert.Len(t, s.names, 10)
		s.Warn("sync: skipped %s in %s (%s)", "file", "Account\nName")
		require.Len(t, hook.AllEntries(), 1)
		msg := hook.LastEntry().Message
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
		assert.True(t, strings.HasPrefix(msg, "sync: skipped 12 files in "), msg)
		assert.NotContains(t, msg, "\n")
		assert.Contains(t, msg, "skipped-09.jpg")
		assert.NotContains(t, msg, "skipped-10.jpg")
		assert.True(t, strings.HasSuffix(msg, " and 2 more)"), msg)
	})
}

func TestSyncNameTooLong(t *testing.T) {
	t.Run("Fits", func(t *testing.T) {
		assert.False(t, syncNameTooLong("/photos/"+strings.Repeat("n", fs.NameMaxLen)))
		assert.False(t, syncNameTooLong(remoteNameOfLen("/", entity.RemoteNameMaxLen, ".jpg")))
	})
	t.Run("FileNameTooLong", func(t *testing.T) {
		assert.True(t, syncNameTooLong("/photos/"+strings.Repeat("n", fs.NameMaxLen+1)))
	})
	t.Run("TooLong", func(t *testing.T) {
		assert.True(t, syncNameTooLong(remoteNameOfLen("/", entity.RemoteNameMaxLen+1, ".jpg")))
	})
}
