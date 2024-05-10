package photoprism

import (
	"path/filepath"
	"regexp"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// BackupAlbums creates a YAML file backup of all albums.
func BackupAlbums(backupPath string, force bool) (count int, err error) {
	c := Config()

	if !c.BackupYaml() && !force {
		log.Debugf("backup: album yaml files disabled")
		return count, nil
	}

	albums, queryErr := query.Albums(0, 1000000)

	if queryErr != nil {
		return count, queryErr
	}

	if !fs.PathExists(backupPath) {
		backupPath = c.AlbumsPath()
	}

	for _, a := range albums {
		fileName := a.YamlFileName(backupPath)

		if saveErr := a.SaveAsYaml(fileName); saveErr != nil {
			log.Errorf("album: %s (update yaml)", saveErr)
			err = saveErr
		} else {
			log.Tracef("backup: saved album yaml file %s", clean.Log(filepath.Base(fileName)))
			count++
		}
	}

	return count, err
}

// RestoreAlbums restores all album YAML file backups.
func RestoreAlbums(backupPath string, force bool) (count int, result error) {
	c := Config()

	if !c.BackupYaml() && !force {
		log.Debugf("restore: album yaml files disabled")
		return count, nil
	}

	existing, err := query.Albums(0, 1)

	if err != nil {
		return count, err
	}

	if len(existing) > 0 && !force {
		log.Debugf("restore: album yaml files disabled")
		return count, nil
	}

	if !fs.PathExists(backupPath) {
		backupPath = c.AlbumsPath()
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
