package batch

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// updateAlbumBackups writes YAML snapshots for all albums referenced in the current batch request
// so the on-disk backups stay in sync with newly added or removed photos.
func updateAlbumBackups(values *PhotosForm) {
	if values == nil || values.Albums.Action != ActionUpdate {
		return
	}

	conf := get.Config()

	if conf == nil || !conf.BackupAlbums() {
		return
	}

	backupPath := conf.BackupAlbumsPath()

	if backupPath == "" {
		return
	}

	rawUIDs := values.Albums.GetValuesByActions([]Action{ActionAdd, ActionRemove})

	if len(rawUIDs) == 0 {
		return
	}

	validUIDs := make([]string, 0, len(rawUIDs))

	for _, uid := range rawUIDs {
		if rnd.InvalidUID(uid, entity.AlbumUID) {
			log.Debugf("batch: invalid album uid %s (skip yaml)", clean.Log(uid))
			continue
		}
		validUIDs = append(validUIDs, uid)
	}

	if len(validUIDs) == 0 {
		return
	}

	albums, err := query.AlbumsByUID(validUIDs, true)

	if err != nil {
		log.Warnf("batch: failed to load albums for yaml backup: %s", err)
		return
	}

	for i := range albums {
		album := &albums[i]

		if album == nil {
			log.Debugf("batch: album is nil (update yaml)")
			continue
		}

		if !album.HasID() {
			log.Debugf("batch: album has no ID (update yaml)")
			continue
		}

		if err = album.SaveBackupYaml(backupPath); err != nil {
			log.Warnf("batch: failed to save album backup %s: %s", clean.Log(album.AlbumUID), err)
		}
	}
}
