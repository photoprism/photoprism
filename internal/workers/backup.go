package workers

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism/backup"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// Backup represents a background backup worker, created with NewBackup.
type Backup struct {
	conf     *config.Config
	database func(backupPath, fileName string, toStdOut, force bool, retain int) error
	albums   func(backupPath string, force bool) (int, error)
}

// NewBackup returns a new Backup worker.
func NewBackup(conf *config.Config) *Backup {
	return &Backup{conf: conf, database: backup.Database, albums: backup.Albums}
}

// backupError names the backup steps that failed and wraps their errors.
type backupError struct {
	steps []string
	errs  []error
}

// add records the error of a failed step.
func (e *backupError) add(step string, err error) {
	e.steps = append(e.steps, step)
	e.errs = append(e.errs, err)
}

// Error returns a summary that names the failed steps.
func (e *backupError) Error() string {
	return fmt.Sprintf("%s backup failed", english.WordSeries(e.steps, "and"))
}

// Unwrap returns the errors of the failed steps.
func (e *backupError) Unwrap() []error {
	return e.errs
}

// StartScheduled starts a scheduled run of the backup worker based on the current configuration.
func (w *Backup) StartScheduled() {
	var failed *backupError

	// Failed steps have already been logged by Start.
	if err := w.Start(w.conf.BackupDatabase(), w.conf.BackupAlbums(), true, w.conf.BackupRetain()); err != nil && !errors.As(err, &failed) {
		log.Errorf("scheduler: %s (backup)", err)
	}
}

// Start creates index and album backups based on the current configuration.
func (w *Backup) Start(database, albums bool, force bool, retain int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("backup: %s (worker panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	// Return if no backups should be created.
	if !database && !albums {
		return nil
	}

	// Reset the cached disk usage so a freshly freed disk is detected immediately.
	disk.FlushFree()

	if w.conf.InsufficientStorage() {
		return status.ErrInsufficientStorage
	}

	// Return error if backup worker is already running.
	if err = mutex.BackupWorker.Start(); err != nil {
		return err
	}

	defer mutex.BackupWorker.Stop()

	// Start creating backups.
	start := time.Now()
	failed := &backupError{}

	// Create database backup.
	if database {
		databasePath := w.conf.BackupDatabasePath()

		if dbErr := w.database(databasePath, "", false, force, retain); dbErr != nil {
			log.Errorf("backup: %s (database)", dbErr)
			failed.add("database", dbErr)
		}
	}

	if mutex.BackupWorker.Canceled() {
		return status.ErrCanceled
	}

	// Create albums backup.
	if albums {
		albumsPath := w.conf.BackupAlbumsPath()

		if count, backupErr := w.albums(albumsPath, false); backupErr != nil {
			log.Errorf("backup: %s (albums)", backupErr.Error())
			failed.add("album", backupErr)
		} else if count > 0 {
			log.Infof("backup: saved %s", english.Plural(count, "album backup", "album backups"))
		}
	}

	// Report the failed steps after every requested step ran.
	if len(failed.errs) > 0 {
		log.Errorf("backup: %s", failed)
		return failed
	}

	elapsed := time.Since(start)

	// Log success message.
	log.Infof("backup: completed in %s", elapsed)

	return nil
}
