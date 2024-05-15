package photoprism

import (
	"path/filepath"
	"regexp"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// RestoreAlbums restores all album YAML file backups.
func RestoreAlbums(backupPath string, force bool) (count int, result error) {
	// Make sure only one backup/restore operation is running at a time.
	backupAlbumsMutex.Lock()
	defer backupAlbumsMutex.Unlock()

	c := Config()

	if !c.BackupAlbums() && !force {
		log.Debugf("restore: album metadata backups are disabled")
		return count, nil
	}

	existing, err := query.Albums(0, 1)

	if err != nil {
		return count, err
	}

	if len(existing) > 0 && !force {
		log.Debugf("restore: found existing albums, backup not restored")
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
			log.Errorf("restore: %s in %s", err, clean.Log(filepath.Base(fileName)))
			result = err
		} else if a.AlbumType == "" || len(a.Photos) == 0 && a.AlbumFilter == "" {
			log.Debugf("restore: skipping %s", clean.Log(filepath.Base(fileName)))
		} else if found := a.Find(); found != nil {
			log.Infof("%s: %s already exists", found.AlbumType, clean.Log(found.AlbumTitle))
		} else if err = a.Create(); err != nil {
			log.Errorf("%s: %s in %s", a.AlbumType, err, clean.Log(filepath.Base(fileName)))
		} else {
			count++
		}
	}

	return count, result
}
