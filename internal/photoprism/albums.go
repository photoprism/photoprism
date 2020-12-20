package photoprism

import (
	"path/filepath"
	"regexp"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
)

// BackupAlbums creates a YAML file backup of all albums.
func BackupAlbums(force bool) (count int, result error) {
	c := Config()

	if !c.BackupYaml() && !force {
		log.Debugf("backup: album yaml files disabled")
		return count, nil
	}

	albums, err := query.GetAlbums(0, 9999)

	if err != nil {
		return count, err
	}

	for _, a := range albums {
		fileName := a.YamlFileName(c.AlbumsPath())

		if err := a.SaveAsYaml(fileName); err != nil {
			log.Errorf("album: %s (update yaml)", err)
			result = err
		} else {
			log.Tracef("backup: saved album yaml file %s", txt.Quote(filepath.Base(fileName)))
			count++
		}
	}

	return count, result
}

// RestoreAlbums restores all album YAML file backups.
func RestoreAlbums(force bool) (count int, result error) {
	c := Config()

	if !c.BackupYaml() && !force {
		log.Debugf("restore: album yaml files disabled")
		return count, nil
	}

	existing, err := query.GetAlbums(0, 1)

	if err != nil {
		return count, err
	}

	if len(existing) > 0 && !force {
		log.Debugf("restore: album yaml files disabled")
		return count, nil
	}

	albums, err := filepath.Glob(regexp.QuoteMeta(c.AlbumsPath()) + "/**/*.yml")

	if err != nil {
		return count, err
	}

	if len(albums) == 0 {
		return count, nil
	}

	for _, fileName := range albums {
		a := entity.Album{}

		if err := a.LoadFromYaml(fileName); err != nil {
			log.Errorf("album: %s (load yaml)", err)
			result = err
		} else if err := a.Create(); err != nil {
			log.Warnf("%s: %s already exists", a.AlbumType, txt.Quote(a.AlbumTitle))
		} else {
			count++
		}
	}

	return count, result
}
