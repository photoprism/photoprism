package workers

import (
	"fmt"
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

// RunPurgeExpired runs isolated expiry tasks; add new tasks here, not to the ticker.
func RunPurgeExpired(conf *config.Config) {
	event.Safe(func() { RunPurgeArchives(conf) })
	event.Safe(func() { purgeStaleUploads(conf) })
}

// purgeStaleUploads scans unlocked and removes expired batches between upload requests.
func purgeStaleUploads(conf *config.Config) {
	if !mutex.UserUploads.Load() || conf.ReadOnly() || conf.UploadMaxAge() < 0 {
		return
	}
	mutex.UserUploads.Store(false)
	root := conf.UsersStoragePath()
	if info, err := os.Lstat(root); os.IsNotExist(err) {
		return
	} else if err != nil {
		mutex.UserUploads.Store(true)
		log.Warnf("upload: %s", clean.Error(err))
		return
	} else if !info.IsDir() {
		return
	}
	now, err := uploadStorageTime(root)
	if err != nil {
		mutex.UserUploads.Store(true)
		log.Warnf("upload: expiry scan skipped because the storage time is unknown (%s)", clean.Error(err))
		return
	}
	cutoff := now.Add(-time.Duration(conf.UploadMaxAge()) * time.Second)
	candidates, pending := scanUploadDirs(root, cutoff, 0)
	result := removeExpiredUploads(candidates, cutoff)
	if pending || result.remaining || result.busy {
		mutex.UserUploads.Store(true)
	}
	if result.busy {
		log.Debug("upload: expiry of batches deferred while requests are active")
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

// uploadClockFile names the file in the users storage folder whose modification time is read as now.
const uploadClockFile = ".upload-purge"

// uploadStorageTime returns the current time of the users storage, see storageTime.
var uploadStorageTime = storageTime

// storageTime replaces the clock file in dir and returns its creation time, so the cutoff and the
// batch times come from the storage's clock, which may differ from the host's. The file is created
// anew and renamed into place, so an existing file is never written.
func storageTime(dir string) (time.Time, error) {
	name := filepath.Join(dir, uploadClockFile)
	if info, err := os.Lstat(name); err == nil && !info.Mode().IsRegular() {
		return time.Time{}, fmt.Errorf("%s is not a regular file", uploadClockFile)
	}
	f, err := os.CreateTemp(dir, uploadClockFile+"-*")
	if err != nil {
		return time.Time{}, err
	}
	tmp := f.Name()
	info, err := f.Stat()
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmp, name)
	}
	if err != nil {
		_ = os.Remove(tmp)
		return time.Time{}, err
	}
	return info.ModTime(), nil
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

// uploadOutcome describes what removeExpiredUpload did with a candidate.
type uploadOutcome int

const (
	uploadSkipped uploadOutcome = iota // no longer a batch directory
	uploadBusy                         // a request holds the lifecycle lock
	uploadKept                         // changed since the scan or failed
	uploadRemoved
)

// removeExpiredUploads rechecks and removes each candidate under its own nonblocking exclusive lock.
func removeExpiredUploads(candidates []string, cutoff time.Time) (result uploadPurgeResult) {
	for _, dir := range candidates {
		outcome, err := removeUploadCandidate(dir, cutoff)
		if err != nil {
			result.errors = append(result.errors, err)
		}
		switch outcome {
		case uploadBusy:
			result.busy = true
		case uploadKept:
			result.remaining = true
		case uploadRemoved:
			result.removed = append(result.removed, dir)
		}
	}
	return result
}

// removeUploadCandidate removes one candidate, see removeExpiredUpload.
var removeUploadCandidate = removeExpiredUpload

// removeExpiredUpload removes the batch if it is still expired, unless a request holds the lifecycle lock.
func removeExpiredUpload(dir string, cutoff time.Time) (uploadOutcome, error) {
	if !mutex.UploadBatches.TryLock() {
		return uploadBusy, nil
	}
	defer mutex.UploadBatches.Unlock()
	info, err := os.Lstat(dir)
	switch {
	case os.IsNotExist(err):
		return uploadSkipped, nil
	case err != nil:
		return uploadKept, err
	case !info.IsDir() || info.Mode()&os.ModeSymlink != 0:
		return uploadSkipped, nil
	}
	if stale, err := uploadBatchExpired(dir, cutoff); err != nil || !stale {
		return uploadKept, err
	} else if err = os.RemoveAll(dir); err != nil {
		return uploadKept, err
	}
	return uploadRemoved, nil
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
