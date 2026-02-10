package query

import (
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/dirs"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// FoldersByPath returns a slice of folders in a given directory incl subfolders in recursive mode with the number of files found populated in FileCount.
func FoldersByPath(rootName, rootPath, path string, recursive, includeRoot bool) (folders entity.Folders, err error) {
	dirs, counts, err := dirs.Dirs(filepath.Join(rootPath, path), recursive, true)

	// Failed?
	if err != nil {
		if len(dirs) == 0 {
			return folders, err
		}
		// At least one folder found.
		log.Infof("folders: %s", err)
	}

	if includeRoot {
		folders = make(entity.Folders, len(dirs))
	} else {
		folders = make(entity.Folders, len(dirs)-1)
	}

	i := 0
	for _, dir := range dirs {
		if (dir == "/" && includeRoot) || (dir != "/") {
			newFolder := entity.NewFolder(rootName, filepath.Join(path, dir), fs.ModTime(filepath.Join(rootPath, dir)))

			if err = newFolder.Create(); err == nil {
				newFolder.FileCount = counts[dir]
				folders[i] = newFolder
				i++
			} else if folder := entity.FindFolder(rootName, filepath.Join(path, dir)); folder != nil {
				folder.FileCount = counts[dir]
				folders[i] = *folder
				i++
			} else {
				log.Errorf("folders: %s (create folder)", err)
			}
		}
	}

	return folders, nil
}

// FolderCoverByUID returns a folder cover file based on the uid.
func FolderCoverByUID(uid string) (file entity.File, err error) {
	if err = Db().Where("files.file_primary = 1 AND files.file_missing = 0 AND files.file_type IN (?) AND files.deleted_at IS NULL", media.PreviewExpr).
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.deleted_at IS NULL AND photos.photo_quality > -1 AND photos.photo_private = 0").
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
	db := UnscopedDb().Table("folders").
		Select("folders.path, folders.root, folders.folder_uid, folders.folder_title, folders.folder_country, folders.folder_year, folders.folder_month, COUNT(photos.id) AS photo_count").
		Joins("JOIN photos ON photos.photo_path = folders.path AND photos.deleted_at IS NULL AND photos.photo_quality >= 3 AND photos.photo_private = 0").
		Group("folders.path, folders.root, folders.folder_uid, folders.folder_title, folders.folder_country, folders.folder_year, folders.folder_month").
		Having("photo_count >= ?", threshold)

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
	case MySQL:
		return UnscopedDb().Exec(`UPDATE folders
		INNER JOIN
			(SELECT photo_path, MAX(taken_at_local) AS taken_max
			FROM photos WHERE taken_src = 'meta' AND photos.photo_quality >= 3 AND photos.deleted_at IS NULL
			GROUP BY photo_path) AS p ON folders.path = p.photo_path
		SET folders.folder_year = YEAR(taken_max), folders.folder_month = MONTH(taken_max), folders.folder_day = DAY(taken_max)
		WHERE p.taken_max IS NOT NULL`).Error
	default:
		return nil
	}
}
