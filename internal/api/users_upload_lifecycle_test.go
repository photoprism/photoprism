package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// uploadLifecycleHook checks request coordination at an observed filesystem outcome.
type uploadLifecycleHook struct {
	prefix     string
	seen       bool
	panicAfter bool
	t          *testing.T
}

// Levels observes upload filesystem outcome logs.
func (h *uploadLifecycleHook) Levels() []logrus.Level { return logrus.AllLevels }

// Fire checks that the request holds a shared lifecycle lock through its filesystem work.
func (h *uploadLifecycleHook) Fire(entry *logrus.Entry) error {
	if h.seen || !strings.Contains(entry.Message, h.prefix) {
		return nil
	}
	h.seen = true
	exclusive := mutex.UploadBatches.TryLock()
	if exclusive {
		mutex.UploadBatches.Unlock()
	}
	assert.False(h.t, exclusive, "request must exclude expiry")
	shared := mutex.UploadBatches.TryRLock()
	if shared {
		mutex.UploadBatches.RUnlock()
	}
	assert.True(h.t, shared, "requests must admit other requests")
	if h.panicAfter {
		panic("upload lifecycle test")
	}
	return nil
}

// TestUploadRequestLifecycle checks shared holds and release on success, failure, and panic.
func TestUploadRequestLifecycle(t *testing.T) {
	for _, tc := range []struct {
		name, method, marker    string
		folderError, panicAfter bool
	}{
		{"Upload", http.MethodPost, "library: uploaded", false, false},
		{"Process", http.MethodPut, "deleted empty folder", false, false},
		{"UploadFolderError", http.MethodPost, "failed to create storage folder", true, false},
		{"ProcessFolderError", http.MethodPut, "failed to access storage folder", true, false},
		{"UploadPanic", http.MethodPost, "upload: saved", false, true},
		{"ProcessPanic", http.MethodPut, "deleted empty folder", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app, router, conf := NewApiTest()
			options := *conf.Options()
			savedLog, savedSystem, flag := log, event.SystemLog, mutex.UserUploads.Load()
			t.Cleanup(func() {
				*conf.Options() = options
				log = savedLog
				event.SystemLog = savedSystem
				mutex.UserUploads.Store(flag)
			})
			conf.Options().ReadOnly = false
			conf.Options().UploadNSFW = true
			conf.Options().UploadAllow = "jpg"
			UploadUserFiles(router)
			ProcessUserUpload(router)
			token := AuthenticateAdmin(app, router)
			name := "lifecycle" + strings.ToLower(tc.name)
			if tc.folderError {
				name = strings.Repeat("t", 300)
			}
			defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(entity.Admin.UserUID), fs.UploadDir), name)
			body := bytes.NewBufferString(`{"albums":[]}`)
			contentType := "application/json"
			// An upload request without files stages the empty batch that the processing request expects.
			if tc.method == http.MethodPut && !tc.folderError {
				staging, stagingType, err := buildMultipart(map[string][]byte{})
				require.NoError(t, err)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+entity.Admin.UserUID+"/upload/"+name, staging)
				req.Header.Set("Content-Type", stagingType)
				header.SetAuthorization(req, token)
				response := httptest.NewRecorder()
				app.ServeHTTP(response, req)
				require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			}
			if tc.method == http.MethodPost {
				var err error
				body, contentType, err = buildMultipart(map[string][]byte{"photo.jpg": NewTestJpeg(t, 160, 160)})
				require.NoError(t, err)
			}
			hook := &uploadLifecycleHook{prefix: tc.marker, panicAfter: tc.panicAfter, t: t}
			logger := logrus.New()
			logger.SetLevel(logrus.DebugLevel)
			logger.SetOutput(io.Discard)
			logger.AddHook(hook)
			log, event.SystemLog = logger, logger
			require.NoError(t, mutex.IndexWorker.Start())
			t.Cleanup(mutex.IndexWorker.Stop)
			req := httptest.NewRequest(tc.method, "/api/v1/users/"+entity.Admin.UserUID+"/upload/"+name, body)
			req.Header.Set("Content-Type", contentType)
			header.SetAuthorization(req, token)
			response := httptest.NewRecorder()
			if tc.panicAfter {
				assert.Panics(t, func() { app.ServeHTTP(response, req) })
			} else {
				app.ServeHTTP(response, req)
				expected := http.StatusOK
				if tc.folderError {
					expected = http.StatusBadRequest
				}
				assert.Equal(t, expected, response.Code, response.Body.String())
			}
			assert.True(t, hook.seen, "filesystem outcome hook must run")
			assert.True(t, mutex.IndexWorker.Running())
			available := mutex.UploadBatches.TryLock()
			if available {
				mutex.UploadBatches.Unlock()
			}
			assert.True(t, available, "handler must release lifecycle lock")
		})
	}
}

// uploadWaitsForLifecycleLock reports whether an upload request is blocked on the shared lifecycle lock.
func uploadWaitsForLifecycleLock() bool {
	buf := make([]byte, 1<<20)
	stacks := string(buf[:runtime.Stack(buf, true)])
	for _, stack := range strings.Split(stacks, "\n\n") {
		if !strings.Contains(stack, "[sync.RWMutex.RLock") {
			continue
		}
		lines := strings.Split(stack, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "sync.(*RWMutex).RLock(") && i+2 < len(lines) {
				if strings.Contains(lines[i+2], "api.UploadUserFiles.func1(") {
					return true
				}
				break
			}
		}
	}
	return false
}

// TestUploadUserFilesBatchUnderLock checks that a request creates its batch folder only while holding the lifecycle lock.
func TestUploadUserFilesBatchUnderLock(t *testing.T) {
	app, router, conf := NewApiTest()
	options, flag := *conf.Options(), mutex.UserUploads.Load()
	name := "lifecyclebatchlock"
	base := filepath.Join(conf.UserStoragePath(entity.Admin.UserUID), fs.UploadDir)
	t.Cleanup(func() {
		*conf.Options() = options
		mutex.UserUploads.Store(flag)
		removeUploadDirsForToken(t, base, name)
	})
	conf.Options().ReadOnly = false
	conf.Options().UploadNSFW = true
	conf.Options().UploadAllow = "jpg"
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)
	batchExists := func() bool {
		matches, err := filepath.Glob(filepath.Join(base, "*"+name))
		require.NoError(t, err)
		return len(matches) > 0
	}
	require.False(t, batchExists())
	body, contentType, err := buildMultipart(map[string][]byte{"photo.jpg": NewTestJpeg(t, 160, 160)})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+entity.Admin.UserUID+"/upload/"+name, body)
	req.Header.Set("Content-Type", contentType)
	header.SetAuthorization(req, token)
	done := make(chan *httptest.ResponseRecorder, 1)
	mutex.UploadBatches.Lock()
	locked := true
	unlock := func() {
		if locked {
			locked = false
			mutex.UploadBatches.Unlock()
		}
	}
	// Runs before the cleanup above, so the request finishes before options and folders are restored.
	t.Cleanup(func() {
		unlock()
		select {
		case <-done:
		case <-time.After(30 * time.Second):
			t.Error("upload request did not finish")
		}
	})
	var finished atomic.Bool
	go func() {
		response := httptest.NewRecorder()
		app.ServeHTTP(response, req)
		finished.Store(true)
		done <- response
	}()
	require.Eventually(t, func() bool { return finished.Load() || uploadWaitsForLifecycleLock() }, 10*time.Second, time.Millisecond)
	if finished.Load() {
		early := <-done
		done <- early
		require.Fail(t, "request finished while the lifecycle lock was held", early.Body.String())
	}
	assert.False(t, batchExists(), "batch folder must not exist before the request holds the lifecycle lock")
	unlock()
	var response *httptest.ResponseRecorder
	select {
	case response = <-done:
		done <- response
	case <-time.After(30 * time.Second):
		require.Fail(t, "upload request did not finish")
	}
	assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.True(t, batchExists())
}

// TestUploadProcessingDuringIndexing imports a staged image while an unrelated worker is active.
func TestUploadProcessingDuringIndexing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	app, router, conf := NewApiTest()
	options, mode, flag := *conf.Options(), conf.AuthMode(), mutex.UserUploads.Load()
	t.Cleanup(func() { *conf.Options() = options; conf.SetAuthMode(mode); mutex.UserUploads.Store(flag) })
	conf.SetAuthMode(config.AuthModePasswd)
	conf.Options().StoragePath = t.TempDir()
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	conf.Options().ImportAllow = ""
	conf.Options().UploadNSFW = true
	conf.Options().UploadAllow = "jpg"
	UploadUserFiles(router)
	ProcessUserUpload(router)
	user := entity.UserFixtures.Pointer("alice")
	session := clientCredentialSession(t, conf, "client", "*", user)
	token := rnd.Base36(10)
	path := "/api/v1/users/" + user.UserUID + "/upload/" + token
	body, contentType, err := buildMultipart(map[string][]byte{"upload.jpg": NewTestJpeg(t, 197, 131)})
	require.NoError(t, err)
	require.NoError(t, mutex.IndexWorker.Start())
	t.Cleanup(mutex.IndexWorker.Stop)
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", contentType)
	header.SetAuthorization(req, session.AuthToken())
	response := httptest.NewRecorder()
	app.ServeHTTP(response, req)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	filename := filepath.Join(conf.UsersStoragePath(), user.UserUID, fs.UploadDir, uploadBatchName(session, token), "upload.jpg")
	require.FileExists(t, filename)
	hash := fs.Hash(filename)
	t.Cleanup(func() {
		file, err := entity.FirstFileByHash(hash)
		if err != nil {
			return
		}
		entity.UnscopedDb().Unscoped().Delete(&entity.PhotoAlbum{}, "photo_uid = ?", file.PhotoUID)
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", file.PhotoID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", file.PhotoID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", file.PhotoID)
	})
	logger, ok := event.Log.(*logrus.Logger)
	require.True(t, ok)
	savedHooks := logger.ReplaceHooks(make(logrus.LevelHooks))
	t.Cleanup(func() { logger.ReplaceHooks(savedHooks) })
	hook := &uploadLifecycleHook{prefix: "import: moving main", t: t}
	logger.AddHook(hook)
	result := AuthenticatedRequestWithBody(app, http.MethodPut, path, `{}`, session.AuthToken())
	require.Equal(t, http.StatusOK, result.Code, result.Body.String())
	require.True(t, hook.seen, "nonempty import must run while holding lifecycle lock")
	file, err := entity.FirstFileByHash(hash)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(conf.OriginalsPath(), file.FileName))
	_, err = os.Stat(filename)
	assert.True(t, os.IsNotExist(err))
	assert.True(t, mutex.IndexWorker.Running())
}
