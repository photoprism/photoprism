package workers

import (
	"bytes"
	iofs "io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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

// newUploadConfig returns a test config that removes staged uploads after the minimum age of one day.
func newUploadConfig(t *testing.T) *config.Config {
	t.Helper()
	c := config.NewMinimalTestConfig(t.TempDir())
	c.Options().UploadMaxAge = config.MinUploadMaxAge
	return c
}

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
		c := newUploadConfig(t)
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
		c := newUploadConfig(t)
		batch := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		purgeStaleUploads(c)
		assert.DirExists(t, batch)
		assert.NoFileExists(t, filepath.Join(c.UsersStoragePath(), uploadClockFile))
		assert.False(t, mutex.UserUploads.Load())
	})
	t.Run("ReadOnly", func(t *testing.T) {
		c := newUploadConfig(t)
		batch := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		c.Options().ReadOnly = true
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		assert.DirExists(t, batch)
		assert.NoFileExists(t, filepath.Join(c.UsersStoragePath(), uploadClockFile))
		assert.True(t, mutex.UserUploads.Load())
	})
	t.Run("IndexingDoesNotDelayExpiry", func(t *testing.T) {
		c := newUploadConfig(t)
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
	t.Run("SetAsideLeftovers", func(t *testing.T) {
		c := newUploadConfig(t)
		savedLog := log
		t.Cleanup(func() { log = savedLog })
		logger := logrus.New()
		var output bytes.Buffer
		logger.SetOutput(&output)
		log = logger
		upload := filepath.Join(c.UsersStoragePath(), "utfrd9md4cywhp5v", fs.UploadDir)
		leftover := filepath.Join(upload, expiredUploadPrefix+"123456")
		require.NoError(t, os.MkdirAll(filepath.Join(leftover, "batch"), fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(leftover, "batch", "photo.jpg"), []byte("photo"), fs.ModeFile))
		expired := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		fresh := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "fresh", time.Hour)
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		assert.NoDirExists(t, leftover)
		assert.NoDirExists(t, expired)
		assert.DirExists(t, fresh)
		entries, err := os.ReadDir(upload)
		require.NoError(t, err)
		assert.Len(t, entries, 1, "only the fresh batch remains")
		assert.Contains(t, output.String(), "upload: removed 1 expired batch")
		assert.Contains(t, output.String(), "upload: removed 2 staged files never imported")
		assert.NotContains(t, output.String(), c.UsersStoragePath())
	})
	t.Run("MissingRoot", func(t *testing.T) {
		c := newUploadConfig(t)
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		assert.False(t, mutex.UserUploads.Load())
		assert.NoDirExists(t, c.UsersStoragePath())
	})
	t.Run("RootError", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("requires filesystem permission enforcement")
		}
		c := newUploadConfig(t)
		batch := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		parent := filepath.Dir(c.UsersStoragePath())
		require.NoError(t, os.Chmod(parent, 0))
		t.Cleanup(func() { require.NoError(t, os.Chmod(parent, 0o700)) }) //nolint:gosec // G302: test directory mode
		savedLog := log
		t.Cleanup(func() { log = savedLog })
		logger := logrus.New()
		var output bytes.Buffer
		logger.SetOutput(&output)
		log = logger
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		require.NoError(t, os.Chmod(parent, 0o700)) //nolint:gosec // G302: test directory mode
		assert.DirExists(t, batch)
		assert.True(t, mutex.UserUploads.Load())
		assert.Contains(t, output.String(), "upload:")
		assert.NotContains(t, output.String(), c.UsersStoragePath())
	})
}

// TestPurgeStaleUploadsMaxAge verifies that the configured maximum age decides which batches expire.
func TestPurgeStaleUploadsMaxAge(t *testing.T) {
	flag, savedLog := mutex.UserUploads.Load(), log
	t.Cleanup(func() { mutex.UserUploads.Store(flag); log = savedLog })
	const user = "utfrd9md4cywhp5v"
	for _, tc := range []struct {
		name   string
		maxAge int64
		window time.Duration
	}{
		{"Default", 0, 7 * 24 * time.Hour},
		{"Minimum", 3600, 24 * time.Hour},
		{"Custom", 3 * 86400, 3 * 24 * time.Hour},
		{"Maximum", 9999999999, 100 * 365 * 24 * time.Hour},
		{"MaxInt64", math.MaxInt64, 100 * 365 * 24 * time.Hour},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := config.NewMinimalTestConfig(t.TempDir())
			c.Options().UploadMaxAge = tc.maxAge
			kept := newUploadBatch(t, c.UsersStoragePath(), user, "within", tc.window-time.Minute)
			removed := newUploadBatch(t, c.UsersStoragePath(), user, "beyond", tc.window+time.Minute)
			mutex.UserUploads.Store(true)
			purgeStaleUploads(c)
			assert.DirExists(t, kept)
			assert.NoDirExists(t, removed)
			assert.True(t, mutex.UserUploads.Load())
		})
	}
	t.Run("Disabled", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())
		c.Options().UploadMaxAge = -1
		batch := newUploadBatch(t, c.UsersStoragePath(), user, "old", 400*24*time.Hour)
		// A folder that cannot be read makes any scan log a warning, unless permissions are not enforced.
		blocked := filepath.Join(c.UsersStoragePath(), "zzblocked")
		require.NoError(t, os.MkdirAll(blocked, fs.ModeDir))
		require.NoError(t, os.Chmod(blocked, 0))
		t.Cleanup(func() { _ = os.Chmod(blocked, fs.ModeDir) })
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		var output bytes.Buffer
		logger.SetOutput(&output)
		log = logger
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		assert.DirExists(t, batch)
		assert.NoFileExists(t, filepath.Join(c.UsersStoragePath(), uploadClockFile))
		assert.True(t, mutex.UserUploads.Load())
		assert.Empty(t, output.String())
	})
}

// TestPurgeStaleUploadsStorageClock verifies that batch age is measured with the storage's clock.
func TestPurgeStaleUploadsStorageClock(t *testing.T) {
	flag, savedClock := mutex.UserUploads.Load(), uploadStorageTime
	t.Cleanup(func() { mutex.UserUploads.Store(flag); uploadStorageTime = savedClock })
	const user = "utfrd9md4cywhp5v"
	for _, tc := range []struct {
		name string
		skew time.Duration
	}{{"Behind", -72 * time.Hour}, {"Ahead", 72 * time.Hour}} {
		t.Run(tc.name, func(t *testing.T) {
			c := newUploadConfig(t)
			uploadStorageTime = func(dir string) (time.Time, error) {
				now, err := storageTime(dir)
				return now.Add(tc.skew), err
			}
			written := newUploadBatch(t, c.UsersStoragePath(), user, "written", -tc.skew)
			expired := newUploadBatch(t, c.UsersStoragePath(), user, "expired", 25*time.Hour-tc.skew)
			mutex.UserUploads.Store(true)
			purgeStaleUploads(c)
			assert.DirExists(t, written)
			assert.NoDirExists(t, expired)
			assert.FileExists(t, filepath.Join(c.UsersStoragePath(), uploadClockFile))
		})
	}
	uploadStorageTime = savedClock
	t.Run("Unusable", func(t *testing.T) {
		for _, kind := range []string{"Symlink", "Dangling", "Directory", "Fifo", "Unwritable"} {
			t.Run(kind, func(t *testing.T) {
				if kind == "Unwritable" && os.Geteuid() == 0 {
					t.Skip("requires filesystem permission enforcement")
				}
				c := newUploadConfig(t)
				batch := newUploadBatch(t, c.UsersStoragePath(), user, "old", 48*time.Hour)
				clock := filepath.Join(c.UsersStoragePath(), uploadClockFile)
				target := filepath.Join(t.TempDir(), "target")
				switch kind {
				case "Symlink":
					require.NoError(t, os.WriteFile(target, []byte("keep"), fs.ModeFile))
					require.NoError(t, os.Symlink(target, clock))
				case "Dangling":
					require.NoError(t, os.Symlink(target, clock))
				case "Directory":
					require.NoError(t, os.Mkdir(clock, fs.ModeDir))
				case "Fifo":
					require.NoError(t, unix.Mkfifo(clock, 0o600))
				case "Unwritable":
					require.NoError(t, os.Chmod(c.UsersStoragePath(), 0o500))                       //nolint:gosec // G302: test directory mode
					t.Cleanup(func() { require.NoError(t, os.Chmod(c.UsersStoragePath(), 0o700)) }) //nolint:gosec // G302: test directory mode
				}
				savedLog := log
				t.Cleanup(func() { log = savedLog })
				logger := logrus.New()
				var output bytes.Buffer
				logger.SetOutput(&output)
				log = logger
				mutex.UserUploads.Store(true)
				purgeStaleUploads(c)
				assert.DirExists(t, batch)
				assert.True(t, mutex.UserUploads.Load())
				assert.Contains(t, output.String(), "expiry scan skipped")
				assert.NotContains(t, output.String(), c.UsersStoragePath())
				switch kind {
				case "Symlink":
					data, err := os.ReadFile(target) //nolint:gosec // G304: test-owned path
					require.NoError(t, err)
					assert.Equal(t, "keep", string(data))
				case "Dangling":
					assert.NoFileExists(t, target)
				}
			})
		}
	})
	t.Run("LinkedRoot", func(t *testing.T) {
		c := newUploadConfig(t)
		target := t.TempDir()
		batch := newUploadBatch(t, target, user, "old", 48*time.Hour)
		require.NoError(t, os.MkdirAll(filepath.Dir(c.UsersStoragePath()), fs.ModeDir))
		require.NoError(t, os.Symlink(target, c.UsersStoragePath()))
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		assert.DirExists(t, batch)
		assert.NoFileExists(t, filepath.Join(target, uploadClockFile))
		assert.False(t, mutex.UserUploads.Load())
	})
}

// TestStorageTime verifies that the clock file is replaced rather than written and non-regular names are refused.
func TestStorageTime(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := t.TempDir()
		name := filepath.Join(dir, uploadClockFile)
		linked := filepath.Join(t.TempDir(), "linked")
		require.NoError(t, os.WriteFile(linked, []byte("keep"), fs.ModeFile))
		require.NoError(t, os.Link(linked, name))
		old := time.Now().Add(-72 * time.Hour)
		require.NoError(t, os.Chtimes(name, old, old))
		before := time.Now().Add(-time.Second)
		now, err := storageTime(dir)
		require.NoError(t, err)
		file, err := os.Lstat(name)
		require.NoError(t, err)
		assert.True(t, now.After(before), now)
		assert.True(t, now.Equal(file.ModTime()))
		assert.Zero(t, file.Size())
		data, err := os.ReadFile(linked) //nolint:gosec // G304: test-owned path
		require.NoError(t, err)
		assert.Equal(t, "keep", string(data))
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		assert.Len(t, entries, 1)
	})
	t.Run("MissingFolder", func(t *testing.T) {
		_, err := storageTime(filepath.Join(t.TempDir(), "missing"))
		assert.Error(t, err)
	})
	for _, kind := range []string{"Symlink", "Dangling", "Directory", "Fifo"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			name := filepath.Join(dir, uploadClockFile)
			target := filepath.Join(t.TempDir(), "target")
			switch kind {
			case "Symlink":
				require.NoError(t, os.WriteFile(target, []byte("keep"), fs.ModeFile))
				require.NoError(t, os.Symlink(target, name))
			case "Dangling":
				require.NoError(t, os.Symlink(target, name))
			case "Directory":
				require.NoError(t, os.Mkdir(name, fs.ModeDir))
			case "Fifo":
				require.NoError(t, unix.Mkfifo(name, 0o600))
			}
			before, err := os.Lstat(name)
			require.NoError(t, err)
			_, err = storageTime(dir)
			assert.ErrorContains(t, err, "not a regular file")
			after, err := os.Lstat(name)
			require.NoError(t, err)
			assert.True(t, os.SameFile(before, after))
			entries, err := os.ReadDir(dir)
			require.NoError(t, err)
			assert.Len(t, entries, 1)
		})
	}
}

// TestPurgeStaleUploadsFolderWarnings verifies that links and files in place of user or upload folders are reported.
func TestPurgeStaleUploadsFolderWarnings(t *testing.T) {
	flag, savedLog := mutex.UserUploads.Load(), log
	t.Cleanup(func() { mutex.UserUploads.Store(flag); log = savedLog })
	capture := func(t *testing.T) *bytes.Buffer {
		logger := logrus.New()
		var output bytes.Buffer
		logger.SetOutput(&output)
		log = logger
		return &output
	}
	t.Run("RootNotDirectory", func(t *testing.T) {
		c := newUploadConfig(t)
		require.NoError(t, os.MkdirAll(filepath.Dir(c.UsersStoragePath()), fs.ModeDir))
		require.NoError(t, os.WriteFile(c.UsersStoragePath(), []byte("file"), fs.ModeFile))
		output := capture(t)
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		assert.Contains(t, output.String(), "users storage folder is a link or not a directory")
		assert.False(t, mutex.UserUploads.Load())
	})
	t.Run("RootLink", func(t *testing.T) {
		c := newUploadConfig(t)
		require.NoError(t, os.MkdirAll(filepath.Dir(c.UsersStoragePath()), fs.ModeDir))
		require.NoError(t, os.Symlink(t.TempDir(), c.UsersStoragePath()))
		output := capture(t)
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		assert.Contains(t, output.String(), "users storage folder is a link or not a directory")
		assert.NotContains(t, output.String(), c.UsersStoragePath())
		assert.False(t, mutex.UserUploads.Load())
	})
	t.Run("LinksAndFiles", func(t *testing.T) {
		c := newUploadConfig(t)
		root := c.UsersStoragePath()
		batch := newUploadBatch(t, root, "utfrd9md4cywhp5v", "old", 48*time.Hour)
		require.NoError(t, os.Symlink(t.TempDir(), filepath.Join(root, "utly13tubtpxbzaz")))
		require.NoError(t, os.MkdirAll(filepath.Join(root, "utmtr7p8m8f8f9ab"), fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(root, "utmtr7p8m8f8f9ab", fs.UploadDir), []byte("file"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(root, "notes.txt"), []byte("file"), fs.ModeFile))
		output := capture(t)
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		assert.NoDirExists(t, batch)
		assert.Contains(t, output.String(), "upload: skipped utly13tubtpxbzaz because it is a link or not a directory")
		assert.Contains(t, output.String(), "upload: skipped utmtr7p8m8f8f9ab/upload because it is a link or not a directory")
		assert.NotContains(t, output.String(), "notes.txt")
		assert.NotContains(t, output.String(), uploadClockFile)
		assert.NotContains(t, output.String(), root)
	})
	t.Run("StaleClockFile", func(t *testing.T) {
		c := newUploadConfig(t)
		newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "fresh", time.Hour)
		stale := filepath.Join(c.UsersStoragePath(), uploadClockFile+"-1")
		require.NoError(t, os.WriteFile(stale, nil, fs.ModeFile))
		ageUploadTree(t, stale, time.Now().Add(-48*time.Hour))
		recent := filepath.Join(c.UsersStoragePath(), uploadClockFile+"-2")
		require.NoError(t, os.WriteFile(recent, nil, fs.ModeFile))
		ageUploadTree(t, recent, time.Now().Add(-time.Hour))
		capture(t)
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		assert.NoFileExists(t, stale)
		assert.FileExists(t, recent, "a temporary clock file younger than the window must be kept")
		assert.FileExists(t, filepath.Join(c.UsersStoragePath(), uploadClockFile))
	})
}

// TestRemoveStaleClockFiles verifies that only clock files older than the cutoff are removed.
func TestRemoveStaleClockFiles(t *testing.T) {
	dir := t.TempDir()
	cutoff := time.Now().Add(-24 * time.Hour)
	old := filepath.Join(dir, uploadClockFile+"-1")
	fresh := filepath.Join(dir, uploadClockFile+"-2")
	clock := filepath.Join(dir, uploadClockFile)
	oldDir := filepath.Join(dir, uploadClockFile+"-3")
	for _, name := range []string{old, fresh, clock} {
		require.NoError(t, os.WriteFile(name, nil, fs.ModeFile))
	}
	require.NoError(t, os.Mkdir(oldDir, fs.ModeDir))
	target := filepath.Join(t.TempDir(), "target")
	require.NoError(t, os.WriteFile(target, nil, fs.ModeFile))
	link := filepath.Join(dir, uploadClockFile+"-4")
	require.NoError(t, os.Symlink(target, link))
	for _, name := range []string{old, clock, oldDir, target, link} {
		ageUploadTree(t, name, time.Now().Add(-48*time.Hour))
	}
	removeStaleClockFiles(dir, cutoff)
	assert.NoFileExists(t, old)
	assert.FileExists(t, fresh)
	assert.FileExists(t, clock)
	assert.DirExists(t, oldDir)
	assert.FileExists(t, target)
	_, err := os.Lstat(link)
	assert.NoError(t, err)
}

// TestRemoveStaleClockFilesPattern verifies that the folder path is not read as a pattern.
func TestRemoveStaleClockFilesPattern(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "lib[a]")
	other := filepath.Join(parent, "liba")
	for _, folder := range []string{dir, other} {
		require.NoError(t, os.Mkdir(folder, fs.ModeDir))
		name := filepath.Join(folder, uploadClockFile+"-1")
		require.NoError(t, os.WriteFile(name, nil, fs.ModeFile))
		ageUploadTree(t, name, time.Now().Add(-48*time.Hour))
	}
	removeStaleClockFiles(dir, time.Now().Add(-24*time.Hour))
	assert.NoFileExists(t, filepath.Join(dir, uploadClockFile+"-1"))
	assert.FileExists(t, filepath.Join(other, uploadClockFile+"-1"))
	unbalanced := filepath.Join(parent, "lib[")
	require.NoError(t, os.Mkdir(unbalanced, fs.ModeDir))
	name := filepath.Join(unbalanced, uploadClockFile+"-1")
	require.NoError(t, os.WriteFile(name, nil, fs.ModeFile))
	ageUploadTree(t, name, time.Now().Add(-48*time.Hour))
	removeStaleClockFiles(unbalanced, time.Now().Add(-24*time.Hour))
	assert.NoFileExists(t, name)
}

// TestRemoveStaleClockFilesWarning verifies that a clock file that cannot be removed is reported by name only.
func TestRemoveStaleClockFilesWarning(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("requires filesystem permission enforcement")
	}
	savedLog := log
	t.Cleanup(func() { log = savedLog })
	logger := logrus.New()
	var output bytes.Buffer
	logger.SetOutput(&output)
	log = logger
	dir := t.TempDir()
	name := filepath.Join(dir, uploadClockFile+"-1")
	require.NoError(t, os.WriteFile(name, nil, fs.ModeFile))
	ageUploadTree(t, name, time.Now().Add(-48*time.Hour))
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	removeStaleClockFiles(dir, time.Now().Add(-24*time.Hour))
	assert.FileExists(t, name)
	assert.Contains(t, output.String(), "upload: failed to remove clock file "+uploadClockFile+"-1")
	assert.NotContains(t, output.String(), dir)
}

// TestUploadScanName verifies that only user and upload folder names are returned.
func TestUploadScanName(t *testing.T) {
	root := filepath.Join(t.TempDir(), "users")
	assert.Equal(t, "utfrd9md4cywhp5v", uploadScanName(filepath.Join(root, "utfrd9md4cywhp5v"), 1))
	assert.Equal(t, filepath.Join("utfrd9md4cywhp5v", fs.UploadDir), uploadScanName(filepath.Join(root, "utfrd9md4cywhp5v", fs.UploadDir), 2))
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
	for _, failure := range []string{"Walk", "SetAside"} {
		t.Run(failure, func(t *testing.T) {
			root := t.TempDir()
			controls := newUploadControls(t, root)
			bad := newUploadBatch(t, root, "utfrd9md4cywhp5v", "bad", 48*time.Hour)
			good := newUploadBatch(t, root, "utly13tubtpxbzaz", "good", 48*time.Hour)
			blocked := bad
			mode := os.FileMode(0)
			if failure == "SetAside" {
				blocked = filepath.Dir(bad)
				mode = 0500
			}
			info, err := os.Stat(blocked)
			require.NoError(t, err)
			require.NoError(t, os.Chmod(blocked, mode))
			t.Cleanup(func() { require.NoError(t, os.Chmod(blocked, info.Mode().Perm())) })
			candidates, _, pending := scanUploadDirs(root, time.Now().Add(-24*time.Hour), 0)
			result := removeExpiredUploads(candidates, time.Now().Add(-24*time.Hour), mutex.UploadRequests.Load())
			assert.Len(t, result.removed, 1)
			if failure == "SetAside" {
				assert.Len(t, result.errors, 1)
			}
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
			c := newUploadConfig(t)
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
		c := newUploadConfig(t)
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
	t.Run("ActiveRequestKeepsFlag", func(t *testing.T) {
		c := newUploadConfig(t)
		batch := newUploadBatch(t, c.UsersStoragePath(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		mutex.UploadBatches.RLock()
		mutex.UserUploads.Store(true)
		purgeStaleUploads(c)
		mutex.UploadBatches.RUnlock()
		assert.DirExists(t, batch)
		assert.True(t, mutex.UserUploads.Load())
	})
	t.Run("RequestStartsAfterScan", func(t *testing.T) {
		root := t.TempDir()
		controls := newUploadControls(t, root)
		batch := newUploadBatch(t, root, "utfrd9md4cywhp5v", "old", 48*time.Hour)
		cutoff := time.Now().Add(-24 * time.Hour)
		requests := mutex.UploadRequests.Load()
		candidates, _, _ := scanUploadDirs(root, cutoff, 0)
		require.Contains(t, candidates, batch)
		mutex.BeginUploadRequest()
		result := removeExpiredUploads(candidates, cutoff, requests)
		mutex.EndUploadRequest()
		assert.True(t, result.busy)
		assert.DirExists(t, batch)

		// The request changed nothing, so the batch is checked again and removed.
		result = removeExpiredUploads(candidates, cutoff, requests)
		assert.Contains(t, result.removed, batch)
		assert.Zero(t, result.deferred)
		assert.NoDirExists(t, batch)
		controls()
	})
	t.Run("CompletedUploadRefreshesCandidate", func(t *testing.T) {
		root := t.TempDir()
		batch := newUploadBatch(t, root, "utfrd9md4cywhp5v", "old", 48*time.Hour)
		cutoff := time.Now().Add(-24 * time.Hour)
		requests := mutex.UploadRequests.Load()
		candidates, _, _ := scanUploadDirs(root, cutoff, 0)
		require.Contains(t, candidates, batch)
		mutex.BeginUploadRequest()
		file := filepath.Join(batch, "nested", "new.jpg")
		require.NoError(t, os.WriteFile(file, []byte("new"), fs.ModeFile))
		mutex.EndUploadRequest()
		result := removeExpiredUploads(candidates, cutoff, requests)
		assert.Empty(t, result.removed)
		assert.True(t, result.remaining)
		assert.Equal(t, 1, result.deferred)
		assert.FileExists(t, file)
	})
	t.Run("CandidateBecomesLink", func(t *testing.T) {
		root := t.TempDir()
		batch := newUploadBatch(t, root, "utfrd9md4cywhp5v", "old", 48*time.Hour)
		cutoff := time.Now().Add(-24 * time.Hour)
		requests := mutex.UploadRequests.Load()
		candidates, _, _ := scanUploadDirs(root, cutoff, 0)
		require.Contains(t, candidates, batch)
		target := t.TempDir()
		require.NoError(t, os.RemoveAll(batch))
		require.NoError(t, os.Symlink(target, batch))
		ageUploadTree(t, batch, time.Now().Add(-48*time.Hour))
		result := removeExpiredUploads(candidates, cutoff, requests)
		assert.Empty(t, result.removed)
		_, err := os.Lstat(batch)
		assert.NoError(t, err)
		assert.DirExists(t, target)
	})
}

// TestSetAsideExpiredUpload verifies that each candidate is set aside under its own short hold of
// the lifecycle lock and removed after it is released.
func TestSetAsideExpiredUpload(t *testing.T) {
	cutoff := time.Now().Add(-24 * time.Hour)
	unlocked := func(t *testing.T) {
		t.Helper()
		available := mutex.UploadBatches.TryLock()
		if available {
			mutex.UploadBatches.Unlock()
		}
		assert.True(t, available, "lifecycle lock must be released")
	}
	t.Run("SetAside", func(t *testing.T) {
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		outcome, aside, err := setAsideExpiredUpload(batch, mutex.UploadRequests.Load())
		require.NoError(t, err)
		assert.Equal(t, uploadSetAside, outcome)
		assert.NoDirExists(t, batch)
		assert.Equal(t, filepath.Dir(batch), filepath.Dir(aside))
		assert.True(t, strings.HasPrefix(filepath.Base(aside), expiredUploadPrefix))
		assert.FileExists(t, filepath.Join(aside, "old", "nested", "photo.jpg"))
		unlocked(t)
		files, err := removeSetAside(aside)
		require.NoError(t, err)
		assert.Equal(t, 1, files)
		assert.NoDirExists(t, aside)
	})
	t.Run("Busy", func(t *testing.T) {
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		requests := mutex.UploadRequests.Load()
		mutex.BeginUploadRequest()
		outcome, aside, err := setAsideExpiredUpload(batch, requests)
		mutex.EndUploadRequest()
		require.NoError(t, err)
		assert.Equal(t, uploadBusy, outcome)
		assert.Empty(t, aside)
		assert.DirExists(t, batch)
		unlocked(t)
	})
	t.Run("RequestSinceScan", func(t *testing.T) {
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		requests := mutex.UploadRequests.Load()
		mutex.BeginUploadRequest()
		mutex.EndUploadRequest()
		outcome, _, err := setAsideExpiredUpload(batch, requests)
		require.NoError(t, err)
		assert.Equal(t, uploadChanged, outcome)
		assert.DirExists(t, batch)
		entries, err := os.ReadDir(filepath.Dir(batch))
		require.NoError(t, err)
		assert.Len(t, entries, 1)
		unlocked(t)
	})
	t.Run("RequestActiveAtScan", func(t *testing.T) {
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		mutex.BeginUploadRequest()
		requests := mutex.UploadRequests.Load()
		mutex.EndUploadRequest()
		outcome, _, err := setAsideExpiredUpload(batch, requests)
		require.NoError(t, err)
		assert.Equal(t, uploadChanged, outcome)
		assert.DirExists(t, batch)
		unlocked(t)
	})
	t.Run("RenameFails", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("requires filesystem permission enforcement")
		}
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		require.NoError(t, os.Chmod(batch, 0o500))
		t.Cleanup(func() { require.NoError(t, os.Chmod(batch, 0o700)) })
		outcome, _, err := setAsideExpiredUpload(batch, mutex.UploadRequests.Load())
		assert.Error(t, err)
		assert.Equal(t, uploadKept, outcome)
		assert.DirExists(t, batch)
		entries, err := os.ReadDir(filepath.Dir(batch))
		require.NoError(t, err)
		assert.Len(t, entries, 1, "a failed rename must not leave a set-aside folder")
		unlocked(t)
	})
	t.Run("NoSpace", func(t *testing.T) {
		saved := makeAsideDir
		t.Cleanup(func() { makeAsideDir = saved })
		makeAsideDir = func(string, string) (string, error) {
			return "", &os.PathError{Op: "mkdir", Path: "aside", Err: syscall.ENOSPC}
		}
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		outcome, aside, err := setAsideExpiredUpload(batch, mutex.UploadRequests.Load())
		require.NoError(t, err)
		assert.Equal(t, uploadRemoved, outcome)
		assert.Empty(t, aside)
		assert.NoDirExists(t, batch)
		makeAsideDir = func(string, string) (string, error) {
			return "", &os.PathError{Op: "mkdir", Path: "aside", Err: syscall.EDQUOT}
		}
		quota := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		result := removeExpiredUploads([]string{quota}, cutoff, mutex.UploadRequests.Load())
		assert.Equal(t, []string{quota}, result.removed)
		assert.NoDirExists(t, quota)
		if os.Geteuid() != 0 {
			locked := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
			nested := filepath.Join(locked, "nested")
			require.NoError(t, os.Chmod(nested, 0o500))
			t.Cleanup(func() { _ = os.Chmod(nested, 0o700) })
			outcome, _, err = setAsideExpiredUpload(locked, mutex.UploadRequests.Load())
			assert.Error(t, err)
			assert.Equal(t, uploadKept, outcome)
		}
		makeAsideDir = func(string, string) (string, error) {
			return "", &os.PathError{Op: "mkdir", Path: "aside", Err: syscall.EACCES}
		}
		other := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		outcome, _, err = setAsideExpiredUpload(other, mutex.UploadRequests.Load())
		assert.Error(t, err)
		assert.Equal(t, uploadKept, outcome)
		assert.DirExists(t, other)
		unlocked(t)
	})
	t.Run("RemovalFails", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("requires filesystem permission enforcement")
		}
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		nested := filepath.Join(batch, "nested")
		require.NoError(t, os.Chmod(nested, 0o500))
		t.Cleanup(func() {
			_ = filepath.WalkDir(filepath.Dir(batch), func(name string, entry iofs.DirEntry, err error) error {
				if err == nil && entry.IsDir() {
					_ = os.Chmod(name, 0o700)
				}
				return nil
			})
		})
		result := removeExpiredUploads([]string{batch}, cutoff, mutex.UploadRequests.Load())
		assert.Empty(t, result.removed)
		assert.Zero(t, result.files)
		assert.True(t, result.remaining)
		assert.Len(t, result.errors, 1)
		assert.NoDirExists(t, batch)
		unlocked(t)
	})
	t.Run("RequestDuringRetry", func(t *testing.T) {
		saved := checkUploadBatch
		t.Cleanup(func() { checkUploadBatch = saved })
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		requests := mutex.UploadRequests.Load()
		mutex.BeginUploadRequest()
		mutex.EndUploadRequest()
		file := filepath.Join(batch, "new.jpg")
		checkUploadBatch = func(dir string, cutoff time.Time) (bool, error) {
			stale, err := saved(dir, cutoff)

			// A request that stages a file after the walk passed it and ends before the retry.
			mutex.BeginUploadRequest()
			assert.NoError(t, os.WriteFile(file, []byte("new"), fs.ModeFile))
			mutex.EndUploadRequest()
			return stale, err
		}
		result := removeExpiredUploads([]string{batch}, cutoff, requests)
		assert.Empty(t, result.removed)
		assert.Equal(t, 1, result.deferred)
		assert.FileExists(t, file)
		unlocked(t)
	})
	t.Run("RetryWalkError", func(t *testing.T) {
		saved := checkUploadBatch
		t.Cleanup(func() { checkUploadBatch = saved })
		checkUploadBatch = func(string, time.Time) (bool, error) { return false, os.ErrPermission }
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		requests := mutex.UploadRequests.Load()
		mutex.BeginUploadRequest()
		mutex.EndUploadRequest()
		result := removeExpiredUploads([]string{batch}, cutoff, requests)
		assert.Empty(t, result.removed)
		assert.Len(t, result.errors, 1)
		assert.True(t, result.remaining)
		assert.DirExists(t, batch)
	})
	t.Run("RetryBatchGone", func(t *testing.T) {
		saved := checkUploadBatch
		t.Cleanup(func() { checkUploadBatch = saved })
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		requests := mutex.UploadRequests.Load()
		mutex.BeginUploadRequest()
		mutex.EndUploadRequest()

		// The batch is processed and removed after the first attempt and before the retry.
		checkUploadBatch = func(dir string, cutoff time.Time) (bool, error) {
			assert.NoError(t, os.RemoveAll(dir))
			return saved(dir, cutoff)
		}
		result := removeExpiredUploads([]string{batch}, cutoff, requests)
		assert.Empty(t, result.removed)
		assert.Empty(t, result.errors)
		assert.Zero(t, result.deferred)
		assert.False(t, result.remaining)
	})
	t.Run("RetryEntryGone", func(t *testing.T) {
		saved := checkUploadBatch
		t.Cleanup(func() { checkUploadBatch = saved })
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		requests := mutex.UploadRequests.Load()
		mutex.BeginUploadRequest()
		mutex.EndUploadRequest()

		// A file inside the batch is moved away while the retry walks it.
		checkUploadBatch = func(dir string, cutoff time.Time) (bool, error) {
			return false, &os.PathError{Op: "lstat", Path: filepath.Join(dir, "nested", "photo.jpg"), Err: syscall.ENOENT}
		}
		result := removeExpiredUploads([]string{batch}, cutoff, requests)
		assert.Empty(t, result.removed)
		assert.Empty(t, result.errors)
		assert.Equal(t, 1, result.deferred)
		assert.True(t, result.remaining)
		assert.DirExists(t, batch)
	})
	t.Run("RetryAfterRequest", func(t *testing.T) {
		root := t.TempDir()
		stale := newUploadBatch(t, root, "utfrd9md4cywhp5v", "stale", 48*time.Hour)
		changed := newUploadBatch(t, root, "utly13tubtpxbzaz", "changed", 48*time.Hour)
		requests := mutex.UploadRequests.Load()
		mutex.BeginUploadRequest()
		file := filepath.Join(changed, "nested", "new.jpg")
		require.NoError(t, os.WriteFile(file, []byte("new"), fs.ModeFile))
		mutex.EndUploadRequest()
		result := removeExpiredUploads([]string{stale, changed}, cutoff, requests)
		assert.Equal(t, []string{stale}, result.removed)
		assert.Equal(t, 1, result.deferred)
		assert.True(t, result.remaining)
		assert.NoDirExists(t, stale)
		assert.FileExists(t, file)
		unlocked(t)
	})
	t.Run("Skipped", func(t *testing.T) {
		requests := mutex.UploadRequests.Load()
		outcome, _, err := setAsideExpiredUpload(filepath.Join(t.TempDir(), "missing"), requests)
		require.NoError(t, err)
		assert.Equal(t, uploadSkipped, outcome)
		target := t.TempDir()
		link := filepath.Join(t.TempDir(), "link")
		require.NoError(t, os.Symlink(target, link))
		outcome, _, err = setAsideExpiredUpload(link, requests)
		require.NoError(t, err)
		assert.Equal(t, uploadSkipped, outcome)
		assert.DirExists(t, target)
		_, err = os.Lstat(link)
		assert.NoError(t, err)
		file := filepath.Join(t.TempDir(), "file.jpg")
		require.NoError(t, os.WriteFile(file, []byte("file"), fs.ModeFile))
		outcome, _, err = setAsideExpiredUpload(file, requests)
		require.NoError(t, err)
		assert.Equal(t, uploadSkipped, outcome)
		assert.FileExists(t, file)
		unlocked(t)
	})
	t.Run("Error", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("requires filesystem permission enforcement")
		}
		requests := mutex.UploadRequests.Load()
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		parent := filepath.Dir(batch)
		require.NoError(t, os.Chmod(parent, 0o500))
		t.Cleanup(func() { require.NoError(t, os.Chmod(parent, 0o700)) })
		outcome, _, err := setAsideExpiredUpload(batch, requests)
		assert.Error(t, err)
		assert.Equal(t, uploadKept, outcome)
		assert.DirExists(t, batch)
		unlocked(t)
		hidden := newUploadBatch(t, t.TempDir(), "utly13tubtpxbzaz", "old", 48*time.Hour)
		hiddenParent := filepath.Dir(hidden)
		require.NoError(t, os.Chmod(hiddenParent, 0))
		t.Cleanup(func() { require.NoError(t, os.Chmod(hiddenParent, 0o700)) })
		outcome, _, err = setAsideExpiredUpload(hidden, requests)
		assert.Error(t, err)
		assert.Equal(t, uploadKept, outcome)
		unlocked(t)
	})
	t.Run("EachCandidate", func(t *testing.T) {
		root := t.TempDir()
		first := newUploadBatch(t, root, "utfrd9md4cywhp5v", "first", 48*time.Hour)
		second := newUploadBatch(t, root, "utly13tubtpxbzaz", "second", 48*time.Hour)
		result := removeExpiredUploads([]string{first, filepath.Join(root, "missing"), second}, cutoff, mutex.UploadRequests.Load())
		assert.Equal(t, []string{first, second}, result.removed)
		assert.Equal(t, 2, result.files)
		assert.False(t, result.remaining)
		assert.False(t, result.busy)
		assert.Empty(t, result.errors)
		for _, dir := range []string{first, second} {
			entries, err := os.ReadDir(filepath.Dir(dir))
			require.NoError(t, err)
			assert.Empty(t, entries, "set-aside folders must be removed")
		}
		unlocked(t)
	})
	t.Run("RemovedAfterUnlock", func(t *testing.T) {
		saved := setAsideUpload
		t.Cleanup(func() { setAsideUpload = saved })
		batch := newUploadBatch(t, t.TempDir(), "utfrd9md4cywhp5v", "old", 48*time.Hour)
		var aside string
		setAsideUpload = func(dir string, requests uint64) (uploadOutcome, string, error) {
			outcome, dest, err := saved(dir, requests)
			aside = dest
			if !mutex.UploadBatches.TryRLock() {
				t.Error("lifecycle lock must be released before the removal")
			} else {
				mutex.UploadBatches.RUnlock()
			}
			assert.DirExists(t, dest, "the set-aside folder is removed after the lock is released")
			return outcome, dest, err
		}
		result := removeExpiredUploads([]string{batch}, cutoff, mutex.UploadRequests.Load())
		assert.Equal(t, []string{batch}, result.removed)
		assert.NoDirExists(t, aside)
	})
	t.Run("BusyCandidate", func(t *testing.T) {
		saved := setAsideUpload
		t.Cleanup(func() { setAsideUpload = saved })
		root := t.TempDir()
		var batches []string
		for _, name := range []string{"first", "active", "last"} {
			batches = append(batches, newUploadBatch(t, root, "utfrd9md4cywhp5v", name, 48*time.Hour))
		}
		setAsideUpload = func(dir string, requests uint64) (uploadOutcome, string, error) {
			if dir != batches[1] {
				return saved(dir, requests)
			}
			if !mutex.UploadBatches.TryRLock() {
				t.Error("lifecycle lock must be released between candidates")
				return saved(dir, requests)
			}
			defer mutex.UploadBatches.RUnlock()
			return saved(dir, requests)
		}
		result := removeExpiredUploads(batches, cutoff, mutex.UploadRequests.Load())
		assert.True(t, result.busy)
		assert.Equal(t, []string{batches[0], batches[2]}, result.removed)
		assert.DirExists(t, batches[1])
		unlocked(t)
	})
}

// TestRemoveSetAside verifies that only the files that were removed are counted.
func TestRemoveSetAside(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		aside := filepath.Join(t.TempDir(), expiredUploadPrefix+"1")
		require.NoError(t, os.MkdirAll(filepath.Join(aside, "batch", "nested"), fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(aside, "batch", "a.jpg"), []byte("a"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(aside, "batch", "nested", "b.jpg"), []byte("b"), fs.ModeFile))
		require.NoError(t, os.Symlink(t.TempDir(), filepath.Join(aside, "batch", "link")))
		files, err := removeSetAside(aside)
		require.NoError(t, err)
		assert.Equal(t, 2, files)
		assert.NoDirExists(t, aside)
	})
	t.Run("PartialFailure", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("requires filesystem permission enforcement")
		}
		aside := filepath.Join(t.TempDir(), expiredUploadPrefix+"2")
		locked := filepath.Join(aside, "batch", "locked")
		require.NoError(t, os.MkdirAll(locked, fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(aside, "batch", "a.jpg"), []byte("a"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(locked, "b.jpg"), []byte("b"), fs.ModeFile))
		require.NoError(t, os.WriteFile(filepath.Join(locked, "c.jpg"), []byte("c"), fs.ModeFile))
		require.NoError(t, os.Chmod(locked, 0o500))
		t.Cleanup(func() { _ = os.Chmod(locked, 0o700) })
		files, err := removeSetAside(aside)
		assert.Error(t, err)
		assert.Equal(t, 1, files)
		result := uploadPurgeResult{}
		result.removeLeftovers([]string{aside})
		assert.True(t, result.remaining)
		assert.Len(t, result.errors, 1)
		assert.Zero(t, result.files)
		require.NoError(t, os.Chmod(locked, 0o700))
		result = uploadPurgeResult{}
		result.removeLeftovers([]string{aside})
		assert.False(t, result.remaining)
		assert.Equal(t, 2, result.files)
		assert.NoDirExists(t, aside)
	})
}

// TestScanUploadLeftovers verifies that only set-aside folders inside an upload folder are leftovers.
func TestScanUploadLeftovers(t *testing.T) {
	root := t.TempDir()
	upload := filepath.Join(root, "utfrd9md4cywhp5v", fs.UploadDir)
	leftover := filepath.Join(upload, expiredUploadPrefix+"1")
	require.NoError(t, os.MkdirAll(leftover, fs.ModeDir))
	target := t.TempDir()
	require.NoError(t, os.Symlink(target, filepath.Join(upload, expiredUploadPrefix+"link")))
	require.NoError(t, os.WriteFile(filepath.Join(upload, expiredUploadPrefix+"file"), []byte("file"), fs.ModeFile))
	userLike := filepath.Join(root, expiredUploadPrefix+"user", fs.UploadDir, expiredUploadPrefix+"2")
	require.NoError(t, os.MkdirAll(userLike, fs.ModeDir))
	_, leftovers, _ := scanUploadDirs(root, time.Now().Add(-24*time.Hour), 0)
	assert.ElementsMatch(t, []string{leftover, userLike}, leftovers)
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
			c := newUploadConfig(t)
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
				assert.False(t, mutex.UserUploads.Load(), "flag must be cleared before the scan")
				held = mutex.UploadBatches.TryRLock()
				assert.True(t, held, "initial scan must permit requests")
				if held {
					mutex.UploadRequests.Add(1)
				}
				if complete && held {
					assert.NoError(t, os.WriteFile(file, []byte("new"), fs.ModeFile))
					mutex.EndUploadRequest()
					held = false
				}
			}}
			logger.AddHook(hook)
			log = logger
			mutex.UserUploads.Store(true)
			purgeStaleUploads(c)
			if held {
				mutex.EndUploadRequest()
			}
			require.True(t, hook.called)
			assert.DirExists(t, batch)
			assert.True(t, mutex.UserUploads.Load())
			if complete {
				assert.FileExists(t, file)
			} else {
				assert.Equal(t, 1, strings.Count(output.String(), "expiry of batches deferred"))
			}
		})
	}
}

// TestStartPurgeExclusions checks that Portal and disabled wakeups leave staged batches untouched.
func TestStartPurgeExclusions(t *testing.T) {
	for _, name := range []string{"Portal", "DisabledWakeup"} {
		t.Run(name, func(t *testing.T) {
			c := newUploadConfig(t)
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
