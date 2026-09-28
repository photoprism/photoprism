package workers

import (
	"bytes"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
)

// ageUploadTree sets every entry's time without following links.
func ageUploadTree(t *testing.T, dir string, at time.Time) {
	t.Helper()
	require.NoError(t, filepath.WalkDir(dir, func(name string, entry iofs.DirEntry, err error) error {
		require.NoError(t, err)
		ts := []unix.Timespec{unix.NsecToTimespec(at.UnixNano()), unix.NsecToTimespec(at.UnixNano())}
		return unix.UtimesNanoAt(unix.AT_FDCWD, name, ts, unix.AT_SYMLINK_NOFOLLOW)
	}))
}

// newUploadBatch creates a nested batch with the requested age.
func newUploadBatch(t *testing.T, root, user, batch string, age time.Duration) string {
	t.Helper()
	dir := filepath.Join(root, user, fs.UploadDir, batch)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "nested"), fs.ModeDir))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "nested", "photo.jpg"), []byte("photo"), fs.ModeFile))
	ageUploadTree(t, dir, time.Now().Add(-age))
	return dir
}

// newUploadControls creates entries that must survive each batch-removal scenario.
func newUploadControls(t *testing.T, root string) func() {
	t.Helper()
	const user = "utsy13tubtpxbzaz"
	fresh := newUploadBatch(t, root, user, "fresh", time.Hour)
	deep := newUploadBatch(t, root, user, "deep", 48*time.Hour)
	now := time.Now()
	deepFile := filepath.Join(deep, "nested", "photo.jpg")
	require.NoError(t, os.Chtimes(deepFile, now, now))
	avatar := filepath.Join(root, user, fs.UploadDir, "avatar.jpg")
	require.NoError(t, os.WriteFile(avatar, []byte("avatar"), fs.ModeFile))
	ageUploadTree(t, avatar, now.Add(-48*time.Hour))
	target := t.TempDir()
	targetFile := filepath.Join(target, "keep.jpg")
	require.NoError(t, os.WriteFile(targetFile, []byte("keep"), fs.ModeFile))
	ageUploadTree(t, target, now.Add(-48*time.Hour))
	link := filepath.Join(root, user, fs.UploadDir, "linked")
	require.NoError(t, os.Symlink(target, link))
	ageUploadTree(t, link, now.Add(-48*time.Hour))
	return func() {
		t.Helper()
		assert.DirExists(t, fresh)
		assert.FileExists(t, deepFile)
		assert.FileExists(t, avatar)
		assert.FileExists(t, targetFile)
		info, err := os.Lstat(link)
		if assert.NoError(t, err) {
			assert.NotZero(t, info.Mode()&os.ModeSymlink)
		}
	}
}

// TestPurgeStaleUploads verifies batch eligibility and the worker's scan guards.
func TestPurgeStaleUploads(t *testing.T) {
	flag := mutex.UserUploads.Load()
	t.Cleanup(func() { mutex.UserUploads.Store(flag) })
	t.Run("Eligibility", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())
		root := c.UsersStoragePath()
		outside := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(outside, "keep.jpg"), []byte("keep"), fs.ModeFile))
		var removed, kept []string
		for _, user := range []string{"utfrd9md4cywhp5v", "utly13tubtpxbzaz"} {
			removed = append(removed, newUploadBatch(t, root, user, "expired", 25*time.Hour))
			kept = append(kept, newUploadBatch(t, root, user, "fresh", time.Hour))
			deep := newUploadBatch(t, root, user, "deep", 25*time.Hour)
			fresh := time.Now()
			freshRoot := newUploadBatch(t, root, user, "freshroot", 25*time.Hour)
			require.NoError(t, os.Chtimes(freshRoot, fresh, fresh))
			kept = append(kept, freshRoot)
			kept = append(kept, newUploadBatch(t, root, user, "withinwindow", 24*time.Hour-time.Minute))
			removed = append(removed, newUploadBatch(t, root, user, "beyondwindow", 24*time.Hour+time.Minute))
			require.NoError(t, os.Chtimes(filepath.Join(deep, "nested", "photo.jpg"), fresh, fresh))
			kept = append(kept, deep)
			freshDir := newUploadBatch(t, root, user, "freshdir", 25*time.Hour)
			require.NoError(t, os.Chtimes(filepath.Join(freshDir, "nested"), fresh, fresh))
			kept = append(kept, freshDir)
			avatar := filepath.Join(root, user, fs.UploadDir, "avatar.jpg")
			require.NoError(t, os.WriteFile(avatar, []byte("avatar"), fs.ModeFile))
			old := fresh.Add(-48 * time.Hour)
			require.NoError(t, os.Chtimes(avatar, old, old))
			kept = append(kept, avatar)
			link := filepath.Join(root, user, fs.UploadDir, "linked")
			require.NoError(t, os.Symlink(outside, link))
			ageUploadTree(t, link, old)
			kept = append(kept, link)
			withLink := newUploadBatch(t, root, user, "withlink", 25*time.Hour)
			require.NoError(t, os.Symlink(outside, filepath.Join(withLink, "link")))
			ageUploadTree(t, withLink, old)
			removed = append(removed, withLink)
		}
		linkedUser := filepath.Join(root, "utmtr7p8m8f8f9ab")
		userTarget := filepath.Join(outside, "user")
		userBatch := newUploadBatch(t, outside, "user", "old", 48*time.Hour)
		require.NoError(t, os.Symlink(userTarget, linkedUser))
		kept = append(kept, userBatch)
		outsideBatch := newUploadBatch(t, outside, "utfrd9md4cywhp5v", "old", 48*time.Hour)
		linkedUpload := filepath.Join(root, "utntr7p8m8f8f9ab", fs.UploadDir)
		require.NoError(t, os.MkdirAll(filepath.Dir(linkedUpload), fs.ModeDir))
		require.NoError(t, os.Symlink(filepath.Dir(outsideBatch), linkedUpload))
		kept = append(kept, linkedUser, linkedUpload, outsideBatch, filepath.Join(outside, "keep.jpg"))
		ageUploadTree(t, outside, time.Now().Add(-48*time.Hour))
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		for _, name := range removed {
			_, err := os.Lstat(name)
			assert.True(t, os.IsNotExist(err), "expired batch remains: %s", name)
		}
		for _, name := range kept {
			_, err := os.Lstat(name)
			assert.NoError(t, err, "retained entry missing: %s", name)
		}
		assert.True(t, mutex.UserUploads.Load())
	})
	t.Run("Idle", func(t *testing.T) {
		mutex.UserUploads.Store(false)
		assert.NotPanics(t, func() { purgeStaleUploads(nil) })
		c := config.NewMinimalTestConfig(t.TempDir())
		batch := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		purgeStaleUploads(c)
		assert.DirExists(t, batch)
		assert.False(t, mutex.UserUploads.Load())
	})
	t.Run("ReadOnly", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())
		batch := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		c.Options().ReadOnly = true
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		assert.DirExists(t, batch)
		assert.True(t, mutex.UserUploads.Load())
	})
	t.Run("IndexingDoesNotDelayExpiry", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())
		controls := newUploadControls(t, c.UsersStoragePath())
		batch := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		require.NoError(t, mutex.IndexWorker.Start())
		t.Cleanup(mutex.IndexWorker.Stop)
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		assert.True(t, mutex.IndexWorker.Running())
		mutex.IndexWorker.Stop()
		assert.NoDirExists(t, batch)
		controls()
		assert.True(t, mutex.UserUploads.Load())
		require.NoError(t, os.RemoveAll(filepath.Join(c.UsersStoragePath(), "utsy13tubtpxbzaz")))
		purgeStaleUploads(c)
		assert.False(t, mutex.UserUploads.Load())
	})
	t.Run("MissingRoot", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		assert.False(t, mutex.UserUploads.Load())
	})
}

// TestUploadBatchExpired verifies strict cutoff comparison and traversal errors.
func TestUploadBatchExpired(t *testing.T) {
	cutoff := time.Now().Truncate(time.Second)
	t.Run("Cutoff", func(t *testing.T) {
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "batch", 48*time.Hour)
		ageUploadTree(t, batch, cutoff)
		expired, err := uploadBatchExpired(batch, cutoff)
		require.NoError(t, err)
		assert.False(t, expired)
		ageUploadTree(t, batch, cutoff.Add(-time.Second))
		expired, err = uploadBatchExpired(batch, cutoff)
		require.NoError(t, err)
		assert.True(t, expired)
	})
	t.Run("Missing", func(t *testing.T) {
		expired, err := uploadBatchExpired(filepath.Join(t.TempDir(), "missing"), cutoff)
		assert.Error(t, err)
		assert.False(t, expired)
	})
}

// TestPurgeUploadDirsErrors verifies that failed batches remain eligible for retry.
func TestPurgeUploadDirsErrors(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("requires filesystem permission enforcement")
	}
	for _, failure := range []string{"Walk", "Remove"} {
		t.Run(failure, func(t *testing.T) {
			root := t.TempDir()
			controls := newUploadControls(t, root)
			bad := newUploadBatch(t, root, "utfrd9md4cywhp5v", "bad", 48*time.Hour)
			good := newUploadBatch(t, root, "utly13tubtpxbzaz", "good", 48*time.Hour)
			blocked := bad
			mode := os.FileMode(0)
			if failure == "Remove" {
				blocked = filepath.Dir(bad)
				mode = 0500
			}
			require.NoError(t, os.Chmod(blocked, mode))
			t.Cleanup(func() { require.NoError(t, os.Chmod(blocked, fs.ModeDir)) })
			candidates, pending := scanUploadDirs(root, time.Now().Add(-24*time.Hour), 0)
			result := removeExpiredUploads(candidates, time.Now().Add(-24*time.Hour))
			assert.Len(t, result.removed, 1)
			assert.True(t, pending || result.remaining)
			assert.DirExists(t, bad)
			assert.NoDirExists(t, good)
			controls()
		})
	}
}

// expiryPanicHook injects a logging panic in one expiry task.
type expiryPanicHook struct {
	prefix string
	called bool
	t      *testing.T
}

// Levels selects expiry log events for injection.
func (h *expiryPanicHook) Levels() []logrus.Level { return logrus.AllLevels }

// Fire panics once when the selected task logs a removal.
func (h *expiryPanicHook) Fire(entry *logrus.Entry) error {
	if !h.called && strings.HasPrefix(entry.Message, h.prefix) {
		h.called = true
		available := mutex.UploadBatches.TryLock()
		if available {
			mutex.UploadBatches.Unlock()
		}
		assert.True(h.t, available, "expiry logging must occur after unlock")
		panic("expiry task test")
	}
	return nil
}

// TestRunPurgeExpired verifies archive and upload tasks with isolated panic recovery.
func TestRunPurgeExpired(t *testing.T) {
	archives, uploads := mutex.TempArchives.Load(), mutex.UserUploads.Load()
	originalLog := log
	t.Cleanup(func() {
		mutex.TempArchives.Store(archives)
		mutex.UserUploads.Store(uploads)
		log = originalLog
	})
	for _, tc := range []struct{ name, prefix string }{{"WithoutPanic", ""}, {"ArchivePanic", "download:"}, {"UploadPanic", "upload:"}, {"ActiveUpload", "active"}} {
		t.Run(tc.name, func(t *testing.T) {
			prefix := tc.prefix
			logger := logrus.New()
			var output bytes.Buffer
			logger.SetOutput(&output)
			logger.SetLevel(logrus.DebugLevel)
			hook := &expiryPanicHook{prefix: prefix, t: t}
			if prefix != "" && prefix != "active" {
				logger.AddHook(hook)
			}
			log = logger
			c := config.NewMinimalTestConfig(t.TempDir())
			batch := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
			controls := newUploadControls(t, c.UsersStoragePath())
			zipDir := filepath.Join(c.TempPath(), fs.ZipDir)
			require.NoError(t, os.MkdirAll(zipDir, fs.ModeDir))
			archive := filepath.Join(zipDir, "expiry-task-test.zip")
			require.NoError(t, os.WriteFile(archive, []byte("zip"), fs.ModeFile))
			old := time.Now().Add(-48 * time.Hour)
			require.NoError(t, os.Chtimes(archive, old, old))
			t.Cleanup(func() { _ = os.Remove(archive) })
			mutex.TempArchives.Store(true)
			mutex.UserUploads.Store(true)
			if prefix == "active" {
				mutex.UploadBatches.RLock()
			}
			assert.NotPanics(t, func() { RunPurgeExpired(c) })
			if prefix == "active" {
				mutex.UploadBatches.RUnlock()
			}
			assert.NoFileExists(t, archive)
			if prefix == "active" {
				assert.DirExists(t, batch)
				assert.True(t, mutex.UserUploads.Load())
			} else {
				assert.NoDirExists(t, batch)
			}
			controls()
			assert.NotContains(t, output.String(), c.UsersStoragePath())
			if prefix != "" && prefix != "active" {
				assert.True(t, hook.called)
			}
		})
	}
}

// TestPurgeUploadLifecycle verifies deterministic scan-to-removal interleavings.
func TestPurgeUploadLifecycle(t *testing.T) {
	flag := mutex.UserUploads.Load()
	t.Cleanup(func() { mutex.UserUploads.Store(flag) })
	t.Run("ActiveRequestSkipsTick", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())
		controls := newUploadControls(t, c.UsersStoragePath())
		batch := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		mutex.UploadBatches.RLock()
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		mutex.UploadBatches.RUnlock()
		assert.DirExists(t, batch)
		assert.True(t, mutex.UserUploads.Load())
		purgeStaleUploads(c)
		assert.NoDirExists(t, batch)
		controls()
	})
	t.Run("RequestStartsAfterScan", func(t *testing.T) {
		root := t.TempDir()
		controls := newUploadControls(t, root)
		batch := newUploadBatch(t, root, "utfrd9md4cywhp5v", "old", 48*time.Hour)
		cutoff := time.Now().Add(-24 * time.Hour)
		candidates, _ := scanUploadDirs(root, cutoff, 0)
		require.Contains(t, candidates, batch)
		mutex.UploadBatches.RLock()
		result := removeExpiredUploads(candidates, cutoff)
		mutex.UploadBatches.RUnlock()
		assert.True(t, result.busy)
		assert.DirExists(t, batch)
		result = removeExpiredUploads(candidates, cutoff)
		assert.Contains(t, result.removed, batch)
		assert.NoDirExists(t, batch)
		controls()
	})
	t.Run("CompletedUploadRefreshesCandidate", func(t *testing.T) {
		root := t.TempDir()
		batch := newUploadBatch(t, root, "utfrd9md4cywhp5v", "old", 48*time.Hour)
		cutoff := time.Now().Add(-24 * time.Hour)
		candidates, _ := scanUploadDirs(root, cutoff, 0)
		require.Contains(t, candidates, batch)
		mutex.UploadBatches.RLock()
		file := filepath.Join(batch, "nested", "new.jpg")
		require.NoError(t, os.WriteFile(file, []byte("new"), fs.ModeFile))
		mutex.UploadBatches.RUnlock()
		result := removeExpiredUploads(candidates, cutoff)
		assert.Empty(t, result.removed)
		assert.True(t, result.remaining)
		assert.FileExists(t, file)
	})
	t.Run("CandidateBecomesLink", func(t *testing.T) {
		root := t.TempDir()
		batch := newUploadBatch(t, root, "utfrd9md4cywhp5v", "old", 48*time.Hour)
		cutoff := time.Now().Add(-24 * time.Hour)
		candidates, _ := scanUploadDirs(root, cutoff, 0)
		require.Contains(t, candidates, batch)
		target := t.TempDir()
		require.NoError(t, os.RemoveAll(batch))
		require.NoError(t, os.Symlink(target, batch))
		ageUploadTree(t, batch, time.Now().Add(-48*time.Hour))
		result := removeExpiredUploads(candidates, cutoff)
		assert.Empty(t, result.removed)
		_, err := os.Lstat(batch)
		assert.NoError(t, err)
		assert.DirExists(t, target)
	})
}

// uploadScanHook observes the lifecycle lock at the scan's error-reporting boundary.
type uploadScanHook struct {
	check  func()
	called bool
}

// Levels selects warnings emitted by an unlocked scan.
func (h *uploadScanHook) Levels() []logrus.Level { return []logrus.Level{logrus.WarnLevel} }

// Fire runs the observation once without adding a production test seam.
func (h *uploadScanHook) Fire(entry *logrus.Entry) error {
	if !h.called {
		h.called = true
		h.check()
	}
	return nil
}

// TestPurgeUploadScanUnlocked verifies request activity between discovery and removal.
func TestPurgeUploadScanUnlocked(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("requires filesystem permission enforcement")
	}
	for _, complete := range []bool{false, true} {
		name := "ActiveRequest"
		if complete {
			name = "CompletedUpload"
		}
		t.Run(name, func(t *testing.T) {
			c := config.NewMinimalTestConfig(t.TempDir())
			batch := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
			blocked := filepath.Join(c.UsersStoragePath(), "zzscan")
			require.NoError(t, os.MkdirAll(blocked, fs.ModeDir))
			info, err := os.Stat(blocked)
			require.NoError(t, err)
			require.NoError(t, os.Chmod(blocked, 0))
			t.Cleanup(func() { require.NoError(t, os.Chmod(blocked, info.Mode().Perm())) })
			savedLog, flag := log, mutex.UserUploads.Load()
			t.Cleanup(func() { log = savedLog; mutex.UserUploads.Store(flag) })
			logger := logrus.New()
			logger.SetLevel(logrus.DebugLevel)
			var output bytes.Buffer
			logger.SetOutput(&output)
			held := false
			file := filepath.Join(batch, "nested", "new.jpg")
			hook := &uploadScanHook{check: func() {
				exclusive := mutex.UploadBatches.TryLock()
				if exclusive {
					mutex.UploadBatches.Unlock()
				}
				assert.True(t, exclusive, "initial scan must hold neither side of lifecycle lock")
				held = mutex.UploadBatches.TryRLock()
				assert.True(t, held, "initial scan must permit requests")
				if complete && held {
					assert.NoError(t, os.WriteFile(file, []byte("new"), fs.ModeFile))
					mutex.UploadBatches.RUnlock()
					held = false
				}
			}}
			logger.AddHook(hook)
			log = logger
			mutex.UserUploads.Store(true)
			purgeStaleUploads(c)
			if held {
				mutex.UploadBatches.RUnlock()
			}
			require.True(t, hook.called)
			assert.DirExists(t, batch)
			assert.True(t, mutex.UserUploads.Load())
			if complete {
				assert.FileExists(t, file)
			} else {
				assert.Equal(t, 1, strings.Count(output.String(), "expiry scan deferred"))
			}
		})
	}
}

// TestStartPurgeExclusions checks that Portal and disabled wakeups leave staged batches untouched.
func TestStartPurgeExclusions(t *testing.T) {
	for _, name := range []string{"Portal", "DisabledWakeup"} {
		t.Run(name, func(t *testing.T) {
			c := config.NewMinimalTestConfig(t.TempDir())
			c.Options().JWTRotateDays = -1
			if name == "Portal" {
				c.Options().Edition = config.Portal
				c.Options().NodeRole = "portal"
				c.Options().WakeupInterval = 60
				require.True(t, c.Portal())
			} else {
				c.Options().Unsafe = true
				c.Options().WakeupInterval = 0
			}
			batch := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
			savedLog, savedScheduler, savedJobs, flag := log, Scheduler, Jobs, mutex.UserUploads.Load()
			Jobs = make(map[string]gocron.Job)
			logger := logrus.New()
			var output bytes.Buffer
			logger.SetOutput(&output)
			log = logger
			mutex.UserUploads.Store(true)
			t.Cleanup(func() { log = savedLog; Scheduler = savedScheduler; Jobs = savedJobs; mutex.UserUploads.Store(flag) })
			assert.NotPanics(t, func() { Start(c) })
			if Scheduler != nil && Scheduler != savedScheduler {
				require.NoError(t, Scheduler.Shutdown())
			}
			assert.Contains(t, output.String(), "disabled metadata, share & sync background workers")
			assert.DirExists(t, batch)
			assert.True(t, mutex.UserUploads.Load())
		})
	}
}
