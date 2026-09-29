package workers

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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
		log.Warnf("upload: expiry scan skipped because the users storage folder is a link or not a directory")
		return
	}
	now, err := uploadStorageTime(root)
	if err != nil {
		mutex.UserUploads.Store(true)
		log.Warnf("upload: expiry scan skipped because the storage time is unknown (%s)", clean.Error(err))
		return
	}
	cutoff := now.Add(-time.Duration(conf.UploadMaxAge()) * time.Second)
	removeStaleClockFiles(root, cutoff)
	records := mutex.LoadUploadRecords()
	candidates, leftovers, pending := scanUploadDirs(root, cutoff, 0)
	result := removeExpiredUploads(candidates, cutoff, &records)
	result.removeLeftovers(leftovers)
	if pending || result.remaining || result.busy {
		mutex.UserUploads.Store(true)
	}
	if result.busy {
		log.Debug("upload: expiry of batches deferred while requests are active")
	}
	if result.deferred > 0 {
		log.Debugf("upload: expiry of %s deferred to the next scan", english.Plural(result.deferred, "changed batch", "changed batches"))
	}
	for _, err := range result.errors {
		log.Warnf("upload: %s", clean.Error(err))
	}
	for _, dir := range result.removed {
		event.SystemDebug([]string{"upload", "removed expired batch %s"}, clean.Log(dir))
	}
	if n := len(result.removed); n > 0 {
		log.Infof("upload: removed %s", english.Plural(n, "expired batch", "expired batches"))
	}
	if result.files > 0 {
		log.Warnf("upload: removed %s never imported", english.Plural(result.files, "staged file", "staged files"))
	}
}

// uploadScanName returns the user and upload folder names of a scanned entry, without the storage path.
func uploadScanName(dir string, depth int) string {
	if depth == 2 {
		return filepath.Join(filepath.Base(filepath.Dir(dir)), filepath.Base(dir))
	}
	return filepath.Base(dir)
}

// removeStaleClockFiles removes temporary clock files left in dir by an interrupted run once they are
// older than cutoff. Names are matched by prefix, so the path of dir is never read as a pattern.
func removeStaleClockFiles(dir string, cutoff time.Time) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Warnf("upload: failed to list clock files (%s)", clean.Error(err))
		return
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), uploadClockFile+"-") {
			continue
		}
		name := filepath.Join(dir, entry.Name())
		if info, lstatErr := os.Lstat(name); lstatErr == nil && info.Mode().IsRegular() && info.ModTime().Before(cutoff) {
			if removeErr := os.Remove(name); removeErr != nil {
				log.Warnf("upload: failed to remove clock file %s (%s)", clean.Log(entry.Name()), clean.Error(removeErr))
			}
		}
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

// expiredUploadPrefix starts the name of a folder into which an expired batch is set aside for
// removal. Batch names cannot contain a dot, so no request can address it.
const expiredUploadPrefix = ".expired-"

// scanUploadDirs collects stale batch candidates and set-aside leftovers without following links.
func scanUploadDirs(dir string, cutoff time.Time, depth int) (candidates, leftovers []string, remaining bool) {
	info, err := os.Lstat(dir)
	switch {
	case os.IsNotExist(err):
		return nil, nil, false
	case err != nil:
		log.Warnf("upload: %s", clean.Error(err))
		return nil, nil, true
	case depth == 1 && info.Mode()&os.ModeSymlink != 0, depth == 2 && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0):
		// User and upload folders are created as directories, so a link in place of either, or a file in
		// place of an upload folder, is reported; other files in the users folder are ignored.
		log.Warnf("upload: skipped %s because it is a link or not a directory", clean.Log(uploadScanName(dir, depth)))
		return nil, nil, false
	case !info.IsDir() || info.Mode()&os.ModeSymlink != 0:
		return nil, nil, false
	}
	if depth == 3 {
		if strings.HasPrefix(filepath.Base(dir), expiredUploadPrefix) {
			return nil, []string{dir}, false
		}
		stale, walkErr := uploadBatchExpired(dir, cutoff)
		if walkErr != nil {
			log.Warnf("upload: %s", clean.Error(walkErr))
			return nil, nil, true
		} else if !stale {
			return nil, nil, true
		}
		return []string{dir}, nil, false
	}
	if depth == 1 {
		return scanUploadDirs(filepath.Join(dir, fs.UploadDir), cutoff, depth+1)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Warnf("upload: %s", clean.Error(err))
		return nil, nil, true
	}
	for _, entry := range entries {
		batches, sets, pending := scanUploadDirs(filepath.Join(dir, entry.Name()), cutoff, depth+1)
		candidates = append(candidates, batches...)
		leftovers = append(leftovers, sets...)
		remaining = remaining || pending
	}
	return candidates, leftovers, remaining
}

// uploadPurgeResult records cleanup outcomes for logging after the lifecycle lock is released.
type uploadPurgeResult struct {
	removed   []string
	files     int
	deferred  int
	errors    []error
	remaining bool
	busy      bool
}

// uploadOutcome describes what setAsideExpiredUpload did with a candidate.
type uploadOutcome int

const (
	uploadSkipped uploadOutcome = iota // no longer a batch directory
	uploadBusy                         // a request holds the lifecycle lock
	uploadChanged                      // a request for the batch began or ended since the scan
	uploadKept                         // failed
	uploadSetAside
	uploadRemoved // removed under the lock, as setting it aside failed on a full disk
)

// removeExpiredUploads sets each candidate aside under its own nonblocking exclusive lock and removes
// the set-aside folders after all candidates were handled, so a request waits at most for one rename
// unless the disk is full. A candidate a request may have changed is checked again and retried once.
func removeExpiredUploads(candidates []string, cutoff time.Time, records *mutex.UploadRecords) (result uploadPurgeResult) {
	var asides, dirs []string
	for _, dir := range candidates {
		outcome, aside, err := setAsideUpload(dir, records.Get(filepath.Base(dir)))
		if outcome == uploadChanged {
			// The record is read before the walk, so a request that ends during it is detected.
			current := mutex.UploadRecord(filepath.Base(dir))
			stale, walkErr := checkUploadBatch(dir, cutoff)
			switch {
			case walkErr == nil && stale:
				outcome, aside, err = setAsideUpload(dir, current)
			case errors.Is(walkErr, iofs.ErrNotExist):
				// Only a batch that is gone is skipped; an entry that vanished inside it defers it.
				if _, lstatErr := os.Lstat(dir); os.IsNotExist(lstatErr) {
					outcome = uploadSkipped
				}
			case walkErr != nil:
				result.errors = append(result.errors, walkErr)
			}
		}
		if err != nil {
			result.errors = append(result.errors, err)
		}
		switch outcome {
		case uploadBusy:
			result.busy = true
		case uploadChanged:
			result.deferred++
			result.remaining = true
		case uploadKept:
			result.remaining = true
		case uploadSetAside:
			asides = append(asides, aside)
			dirs = append(dirs, dir)
		case uploadRemoved:
			result.removed = append(result.removed, dir)
		}
	}
	for i, aside := range asides {
		files, err := removeSetAside(aside)
		result.files += files
		if err != nil {
			result.errors = append(result.errors, err)
			result.remaining = true
		} else {
			result.removed = append(result.removed, dirs[i])
		}
	}
	return result
}

// removeLeftovers removes set-aside folders that a previous run could not remove.
func (result *uploadPurgeResult) removeLeftovers(leftovers []string) {
	for _, aside := range leftovers {
		files, err := removeSetAside(aside)
		result.files += files
		if err != nil {
			result.errors = append(result.errors, err)
			result.remaining = true
		} else {
			event.SystemDebug([]string{"upload", "removed set-aside folder %s"}, clean.Log(aside))
		}
	}
}

// checkUploadBatch checks a candidate again before it is retried, see uploadBatchExpired.
var checkUploadBatch = uploadBatchExpired

// setAsideUpload sets one candidate aside, see setAsideExpiredUpload.
var setAsideUpload = setAsideExpiredUpload

// makeAsideDir creates the set-aside folder, see os.MkdirTemp.
var makeAsideDir = os.MkdirTemp

// setAsideExpiredUpload renames an expired batch into a new set-aside folder next to it, unless a
// request holds the lifecycle lock or the batch's request record changed since the given value.
// If the folder cannot be created on a full disk, the batch is removed under the lock instead.
func setAsideExpiredUpload(dir string, record uint64) (outcome uploadOutcome, aside string, err error) {
	if !mutex.UploadBatches.TryLock() {
		return uploadBusy, "", nil
	}
	defer mutex.UploadBatches.Unlock()
	info, err := os.Lstat(dir)
	switch {
	case os.IsNotExist(err):
		return uploadSkipped, "", nil
	case err != nil:
		return uploadKept, "", err
	case !info.IsDir() || info.Mode()&os.ModeSymlink != 0:
		return uploadSkipped, "", nil
	case mutex.UploadRecord(filepath.Base(dir)) != record:
		return uploadChanged, "", nil
	}
	if aside, err = makeAsideDir(filepath.Dir(dir), expiredUploadPrefix); errors.Is(err, syscall.ENOSPC) || errors.Is(err, syscall.EDQUOT) {
		if err = os.RemoveAll(dir); err != nil {
			return uploadKept, "", err
		}
		return uploadRemoved, "", nil
	} else if err != nil {
		return uploadKept, "", err
	} else if err = os.Rename(dir, filepath.Join(aside, filepath.Base(dir))); err != nil {
		_ = os.Remove(aside)
		return uploadKept, "", err
	}
	return uploadSetAside, aside, nil
}

// removeSetAside removes a set-aside folder and returns how many regular files were removed with it.
func removeSetAside(aside string) (files int, err error) {
	files = countUploadFiles(aside)
	if err = os.RemoveAll(aside); err != nil {
		files -= countUploadFiles(aside)
	}
	return files, err
}

// countUploadFiles returns the number of regular files in dir without following links.
func countUploadFiles(dir string) (files int) {
	_ = filepath.WalkDir(dir, func(name string, entry iofs.DirEntry, walkErr error) error {
		if walkErr == nil && entry.Type().IsRegular() {
			files++
		}
		return nil
	})
	return files
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
