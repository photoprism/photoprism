package photoprism

import (
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

var backupAlbumsLatest = time.Time{}
var backupAlbumsMutex = sync.Mutex{}

// BackupAlbums creates a YAML file backup of all albums.
func BackupAlbums(backupPath string, force bool) (count int, err error) {
	// Make sure only one backup/restore operation is running at a time.
	backupAlbumsMutex.Lock()
	defer backupAlbumsMutex.Unlock()

	albums, queryErr := query.Albums(0, 1000000)

	if queryErr != nil {
		return count, queryErr
	}

	if !fs.PathExists(backupPath) {
		backupPath = Config().BackupAlbumsPath()
	}

	log.Infof("exporting album metadata to YAML backup files")
	log.Debugf("album backups will be created in %s", clean.Log(backupPath))

	var latest time.Time

	if !force {
		latest = backupAlbumsLatest
	}

	// Save albums to YAML backup files.
	for _, a := range albums {
		// Skip albums that have already been saved to YAML backup files.
		if !force && !backupAlbumsLatest.IsZero() && !a.UpdatedAt.IsZero() &&
			a.UpdatedAt.Before(backupAlbumsLatest) {
			continue
		}

		// Remember most recent date.
		if a.UpdatedAt.After(latest) {
			latest = a.UpdatedAt
		}

		// Write album metadata to YAML backup file.
		if saveErr := a.SaveBackupYaml(backupPath); saveErr != nil {
			err = saveErr
		} else {
			count++
		}
	}

	backupAlbumsLatest = latest

	return count, err
}
