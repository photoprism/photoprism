package workers

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
)

// Backup represents a background backup worker.
type Backup struct {
	conf *config.Config
}

// NewBackup returns a new Backup worker.
func NewBackup(conf *config.Config) *Backup {
	return &Backup{conf: conf}
}

// StartScheduled starts a scheduled run of the backup worker based on the current configuration.
func (w *Backup) StartScheduled() {
	if err := w.Start(w.conf.BackupDatabase(), w.conf.BackupAlbums(), true, w.conf.BackupRetain()); err != nil {
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

	// Return error if backup worker is already running.
	if err = mutex.BackupWorker.Start(); err != nil {
		return err
	}

	defer mutex.BackupWorker.Stop()

	// Start creating backups.
	start := time.Now()

	// Create database backup.
	if database {
		databasePath := w.conf.BackupDatabasePath()

		if err = photoprism.BackupDatabase(databasePath, "", false, force, retain); err != nil {
			log.Errorf("backup: %s (database)", err)
		}
	}

	if mutex.BackupWorker.Canceled() {
		return errors.New("canceled")
	}

	// Create albums backup.
	if albums {
		albumsPath := w.conf.BackupAlbumsPath()

		if count, backupErr := photoprism.BackupAlbums(albumsPath, false); backupErr != nil {
			log.Errorf("backup: %s (albums)", backupErr.Error())
		} else if count > 0 {
			log.Infof("backup: saved %s", english.Plural(count, "album backup", "album backups"))
		}
	}

	elapsed := time.Since(start)

	// Log success message.
	log.Infof("backup: completed in %s", elapsed)

	return nil
}
