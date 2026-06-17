package query

import (
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// FoldersByPath returns a slice of folders in a given directory incl subfolders in recursive mode.
func FoldersByPath(rootName, rootPath, path string, recursive bool) (folders entity.Folders, err error) {
	dirs, err := fs.Dirs(filepath.Join(rootPath, path), recursive, true)

	// Failed?
	if err != nil {
		if len(dirs) == 0 {
			return folders, err
		} else {
			// At least one folder found.
			log.Infof("folders: %s", err)
		}
	}

	folders = make(entity.Folders, len(dirs))

	for i, dir := range dirs {
		newFolder := entity.NewFolder(rootName, filepath.Join(path, dir), fs.ModTime(filepath.Join(rootPath, dir)))

		result, _, err := entity.FirstOrCreateFolder(&newFolder)
		if err == nil {
			folders[i] = *result
		} else {
			log.Errorf("folders: %s (create folder)", err)
		}
	}

	return folders, nil
}

// FolderCoverByUID returns a folder cover file based on the uid.
func FolderCoverByUID(uid string) (file entity.File, err error) {
	if err = Db().Where("files.file_primary = TRUE AND files.file_missing = FALSE AND files.file_type IN (?) AND files.deleted_at IS NULL", media.PreviewExpr).
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.deleted_at IS NULL AND photos.photo_quality > -1 AND photos.photo_private = FALSE").
		Joins("JOIN folders ON photos.photo_path = folders.path AND folders.folder_uid = ?", uid).
		Order("photos.photo_quality DESC").
		Limit(1).
		First(&file).Error; err != nil {
		return file, err
	}

	return file, nil
}

// AlbumFolders returns folders that should be added as album.
func AlbumFolders(threshold int) (folders entity.Folders, err error) {
	folders = make(entity.Folders, 0)
	db := UnscopedDb().Table("folders").
		Select("folders.path, folders.root, folders.folder_uid, folders.folder_title, folders.folder_country, folders.folder_year, folders.folder_month, COUNT(photos.id) AS photo_count").
		Joins("JOIN photos ON photos.photo_path = folders.path AND photos.deleted_at IS NULL AND photos.photo_quality >= 3 AND photos.photo_private = FALSE").
		Group("folders.path, folders.root, folders.folder_uid, folders.folder_title, folders.folder_country, folders.folder_year, folders.folder_month").
		Having("COUNT(photos.id) >= ?", threshold)

	if err = db.Scan(&folders).Error; err != nil {
		return folders, err
	}

	return folders, nil
}

// UpdateFolderDates updates the year, month and day of the folder based on the indexed photo metadata.
func UpdateFolderDates() error {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	switch DbDialect() {
	case dsn.DialectMySQL:
		return UnscopedDb().Exec(`UPDATE folders
		INNER JOIN
			(SELECT photo_path, MAX(taken_at_local) AS taken_max
			FROM photos WHERE taken_src = 'meta' AND photos.photo_quality >= 3 AND photos.deleted_at IS NULL
			GROUP BY photo_path) AS p ON folders.path = p.photo_path
		SET folders.folder_year = YEAR(taken_max), folders.folder_month = MONTH(taken_max), folders.folder_day = DAY(taken_max)
		WHERE p.taken_max IS NOT NULL`).Error
	case dsn.DialectPostgreSQL:
		return UnscopedDb().Exec(`UPDATE folders
			SET folder_year = date_part('year', taken_max), folder_month = date_part('month', taken_max), folder_day = date_part('day', taken_max)
			FROM (SELECT photo_path, MAX(taken_at_local) AS taken_max
	 			FROM photos WHERE taken_src = 'meta' AND photos.photo_quality >= 3 AND photos.deleted_at IS NULL
	 			GROUP BY photo_path
			) AS p
			WHERE folders.path = p.photo_path AND p.taken_max IS NOT NULL`).Error
	case dsn.DialectSQLite:
		return UnscopedDb().Exec(`UPDATE folders
			SET folder_year = strftime('%Y', taken_max), folder_month = strftime('%m', taken_max), folder_day = strftime('%d', taken_max)
			FROM (SELECT photo_path, MAX(taken_at_local) AS taken_max
	 			FROM photos WHERE taken_src = 'meta' AND photos.photo_quality >= 3 AND photos.deleted_at IS NULL
	 			GROUP BY photo_path
			) AS p
			WHERE folders.path = p.photo_path AND p.taken_max IS NOT NULL`).Error
	default:
		return nil
	}
}
