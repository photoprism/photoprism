package workers

import (
	iofs "io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const staleUploadAge = 24 * time.Hour

// RunPurgeExpired runs isolated expiry tasks; add new tasks here, not to the ticker.
func RunPurgeExpired(conf *config.Config) {
	event.Safe(func() { RunPurgeArchives(conf) })
	event.Safe(func() { purgeStaleUploads(conf) })
}

// purgeStaleUploads scans unlocked and removes expired batches between upload requests.
func purgeStaleUploads(conf *config.Config) {
	if !mutex.UserUploads.Load() || conf.ReadOnly() {
		return
	}
	mutex.UserUploads.Store(false)
	cutoff := time.Now().Add(-staleUploadAge)
	candidates, pending := scanUploadDirs(conf.UsersStoragePath(), cutoff, 0)
	result := removeExpiredUploads(candidates, cutoff)
	if pending || result.remaining || result.busy {
		mutex.UserUploads.Store(true)
	}
	if result.busy {
		log.Debug("upload: expiry scan deferred while requests are active")
	}
	for _, err := range result.errors {
		log.Warnf("upload: %s", clean.Error(err))
	}
	for _, dir := range result.removed {
		event.SystemDebug([]string{"upload", "removed expired batch %s"}, clean.Log(dir))
	}
	if len(result.removed) > 0 {
		log.Infof("upload: removed %s", english.Plural(len(result.removed), "expired batch", "expired batches"))
	}
}

// scanUploadDirs collects stale batch candidates without following directory links.
func scanUploadDirs(dir string, cutoff time.Time, depth int) (candidates []string, remaining bool) {
	info, err := os.Lstat(dir)
	switch {
	case os.IsNotExist(err):
		return nil, false
	case err != nil:
		log.Warnf("upload: %s", clean.Error(err))
		return nil, true
	case !info.IsDir() || info.Mode()&os.ModeSymlink != 0:
		return nil, false
	}
	if depth == 3 {
		stale, walkErr := uploadBatchExpired(dir, cutoff)
		if walkErr != nil {
			log.Warnf("upload: %s", clean.Error(walkErr))
			return nil, true
		} else if !stale {
			return nil, true
		}
		return []string{dir}, false
	}
	if depth == 1 {
		return scanUploadDirs(filepath.Join(dir, fs.UploadDir), cutoff, depth+1)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Warnf("upload: %s", clean.Error(err))
		return nil, true
	}
	for _, entry := range entries {
		batches, pending := scanUploadDirs(filepath.Join(dir, entry.Name()), cutoff, depth+1)
		candidates = append(candidates, batches...)
		remaining = remaining || pending
	}
	return candidates, remaining
}

// uploadPurgeResult records cleanup outcomes for logging after the lifecycle lock is released.
type uploadPurgeResult struct {
	removed   []string
	errors    []error
	remaining bool
	busy      bool
}

// removeExpiredUploads rechecks candidates under a nonblocking exclusive lifecycle lock.
func removeExpiredUploads(candidates []string, cutoff time.Time) (result uploadPurgeResult) {
	if !mutex.UploadBatches.TryLock() {
		result.busy = true
		return result
	}
	defer mutex.UploadBatches.Unlock()
	for _, dir := range candidates {
		info, err := os.Lstat(dir)
		switch {
		case os.IsNotExist(err):
			continue
		case err != nil:
			result.errors = append(result.errors, err)
			result.remaining = true
			continue
		case !info.IsDir() || info.Mode()&os.ModeSymlink != 0:
			continue
		}
		stale, err := uploadBatchExpired(dir, cutoff)
		if err == nil && stale {
			err = os.RemoveAll(dir)
		}
		switch {
		case err != nil:
			result.errors = append(result.errors, err)
			result.remaining = true
		case !stale:
			result.remaining = true
		default:
			result.removed = append(result.removed, dir)
		}
	}
	return result
}

// uploadBatchExpired reports whether every entry, including the batch directory, predates cutoff.
func uploadBatchExpired(dir string, cutoff time.Time) (bool, error) {
	fresh := false
	err := filepath.WalkDir(dir, func(name string, entry iofs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		info, err := os.Lstat(name)
		if err != nil {
			return err
		} else if !info.ModTime().Before(cutoff) {
			fresh = true
			return filepath.SkipAll
		}

		return nil
	})

	return !fresh && err == nil, err
}
