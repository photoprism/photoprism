package workers

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"time"

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
	if err := w.Start(w.conf.BackupIndex(), w.conf.BackupAlbums(), true); err != nil {
		log.Errorf("scheduler: %s (backup)", err)
	}
}

// Start creates index and album backups based on the current configuration.
func (w *Backup) Start(index, albums, force bool) (err error) {
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
	backupPath := filepath.Join(w.conf.BackupPath(), w.conf.DatabaseDriver())

	if index && albums {
		log.Infof("backup: creating index and album backups")
	} else if index {
		log.Infof("backup: creating index backup")
	} else {
		log.Infof("backup: creating album backup")
	}

	// Create index database backup.
	if !index {
		// Skip.
	} else if err = photoprism.BackupIndex(backupPath, "", false, force); err != nil {
		log.Errorf("backup: %s (backup index)", err)
	}

	if mutex.BackupWorker.Canceled() {
		return errors.New("backup: canceled")
	}

	// Create album YAML file backup.
	if !albums {
		// Skip.
	} else if count, backupErr := photoprism.BackupAlbums(w.conf.AlbumsPath(), force); backupErr != nil {
		log.Errorf("backup: %s (backup albums)", backupErr.Error())
	} else if count > 0 {
		log.Debugf("backup: %d albums saved as yaml files", count)
	}

	// Update time when worker was last executed.
	w.lastRun = entity.TimeStamp()

	elapsed := time.Since(start)

	// Show success message.
	log.Infof("backup: completed in %s", elapsed)

	return nil
}
