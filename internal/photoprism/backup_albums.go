package photoprism

import (
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

var backupAlbumsTime = time.Time{}
var backupAlbumsMutex = sync.Mutex{}

// BackupAlbums creates a YAML file backup of all albums.
func BackupAlbums(backupPath string, force bool) (count int, err error) {
	// Make sure only one backup/restore operation is running at a time.
	backupAlbumsMutex.Lock()
	defer backupAlbumsMutex.Unlock()

	// Get albums from database.
	albums, queryErr := query.Albums(0, 1000000)

	if queryErr != nil {
		return count, queryErr
	}

	if !fs.PathExists(backupPath) {
		backupPath = Config().BackupAlbumsPath()
	}

	log.Debugf("backup: album backups will be stored in %s", clean.Log(backupPath))
	log.Infof("backup: saving album metadata in YAML backup files")

	var latest time.Time

	// Ignore the last modification timestamp if the force flag is set.
	if !force {
		latest = backupAlbumsTime
	}

	// Save albums to YAML backup files.
	for _, a := range albums {
		// Album modification timestamp.
		changed := a.UpdatedAt

		// Skip albums that have already been saved to YAML backup files.
		if !force && !backupAlbumsTime.IsZero() && !changed.IsZero() && !backupAlbumsTime.Before(changed) {
			continue
		}

		// Remember the lastest modification timestamp.
		if changed.After(latest) {
			latest = changed
		}

		// Write album metadata to YAML backup file.
		if saveErr := a.SaveBackupYaml(backupPath); saveErr != nil {
			err = saveErr
		} else {
			count++
		}
	}

	// Set backupAlbumsTime to latest modification timestamp,
	// so that already saved albums can be skipped next time.
	backupAlbumsTime = latest

	return count, err
}
