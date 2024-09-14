package backup

import (
	"path/filepath"
	"regexp"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Albums creates a YAML file backup of all albums.
func Albums(backupPath string, force bool) (count int, err error) {
	// Make sure only one backup/restore operation is running at a time.
	backupAlbumsMutex.Lock()
	defer backupAlbumsMutex.Unlock()

	// Get albums from database.
	albums, queryErr := query.Albums(0, 1000000)

	if queryErr != nil {
		return count, queryErr
	}

	if !fs.PathExists(backupPath) {
		backupPath = get.Config().BackupAlbumsPath()
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

// RestoreAlbums restores all album YAML file backups.
func RestoreAlbums(backupPath string, force bool) (count int, result error) {
	// Make sure only one backup/restore operation is running at a time.
	backupAlbumsMutex.Lock()
	defer backupAlbumsMutex.Unlock()

	c := get.Config()

	if !c.BackupAlbums() && !force {
		log.Debugf("albums: metadata backup files are disabled")
		return count, nil
	}

	existing, err := query.Albums(0, 1)

	if err != nil {
		return count, err
	}

	if len(existing) > 0 && !force {
		log.Debugf("albums: skipped restoring backups because albums already exist")
		return count, nil
	}

	if !fs.PathExists(backupPath) {
		backupPath = c.BackupAlbumsPath()
	}

	albums, err := filepath.Glob(regexp.QuoteMeta(backupPath) + "/**/*.yml")

	if oAlbums, oErr := filepath.Glob(regexp.QuoteMeta(c.OriginalsAlbumsPath()) + "/**/*.yml"); oErr == nil {
		err = nil
		albums = append(albums, oAlbums...)
	}

	if err != nil {
		return count, err
	}

	if len(albums) == 0 {
		return count, nil
	}

	for _, fileName := range albums {
		a := entity.Album{}

		if err = a.LoadFromYaml(fileName); err != nil {
			log.Errorf("albums: %s in %s (restore)", err, clean.Log(filepath.Base(fileName)))
			result = err
		} else if a.AlbumType == "" || len(a.Photos) == 0 && a.AlbumFilter == "" {
			log.Debugf("albums: skipped %s (restore)", clean.Log(filepath.Base(fileName)))
		} else if found := a.Find(); found != nil {
			log.Infof("%s: %s already exists (restore)", found.AlbumType, clean.Log(found.AlbumTitle))
		} else if err = a.Create(); err != nil {
			log.Errorf("%s: %s in %s (restore)", a.AlbumType, err, clean.Log(filepath.Base(fileName)))
		} else {
			count++
		}
	}

	return count, result
}
