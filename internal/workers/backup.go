package workers

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
)

// Backup represents a background backup worker.
type Backup struct {
	conf    *config.Config
	lastRun time.Time
}

// NewBackup returns a new Backup worker.
func NewBackup(conf *config.Config) *Backup {
	return &Backup{conf: conf}
}

// StartScheduled starts a scheduled run of the backup worker based on the current configuration.
func (w *Backup) StartScheduled() {
	if err := w.Start(w.conf.BackupIndex(), w.conf.BackupAlbums(), true, w.conf.BackupRetain()); err != nil {
		log.Errorf("scheduler: %s (backup)", err)
	}
}

// Start creates index and album backups based on the current configuration.
func (w *Backup) Start(index, albums bool, force bool, retain int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("backup: %s (worker panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	// Return if no backups should be created.
	if !index && !albums {
		return nil
	}

	// Return error if backup worker is already running.
	if err = mutex.BackupWorker.Start(); err != nil {
		return err
	}

	defer mutex.BackupWorker.Stop()

	// Start creating backups.
	start := time.Now()

	// Create index database backup.
	if index {
		backupPath := w.conf.BackupIndexPath()

		if err = photoprism.BackupIndex(backupPath, "", false, force, retain); err != nil {
			log.Errorf("backup: %s (index)", err)
		}
	}

	if mutex.BackupWorker.Canceled() {
		return errors.New("canceled")
	}

	// Create album YAML file backup.
	if albums {
		albumsBackupPath := w.conf.BackupAlbumsPath()

		if count, backupErr := photoprism.BackupAlbums(albumsBackupPath, false); backupErr != nil {
			log.Errorf("backup: %s (album)", backupErr.Error())
		} else if count > 0 {
			log.Infof("exported %s", english.Plural(count, "album", "albums"))
		}
	}

	// Update time when worker was last executed.
	w.lastRun = entity.TimeStamp()

	elapsed := time.Since(start)

	// Show success message.
	log.Infof("backup: completed in %s", elapsed)

	return nil
}
