package query

import (
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

// FoldersByPath returns a slice of folders in a given directory incl sub directories in recursive mode.
func FoldersByPath(rootName, rootPath, path string, recursive bool) (folders entity.Folders, err error) {
	dirs, err := fs.Dirs(filepath.Join(rootPath, path), recursive, true)

	if err != nil {
		return folders, err
	}

	folders = make(entity.Folders, len(dirs))

	for i, dir := range dirs {
		folder := entity.FindFolder(rootName, filepath.Join(path, dir))

		if folder == nil {
			newFolder := entity.NewFolder(rootName, filepath.Join(path, dir), nil)

			if err := newFolder.Create(); err != nil {
				log.Errorf("folders: %s (create folder)", err.Error())
			} else {
				folders[i] = newFolder
			}
		} else {
			folders[i] = *folder
		}
	}

	return folders, nil
}

// AlbumFolders returns folders that should be added as album.
func AlbumFolders(threshold int) (folders entity.Folders, err error) {
	db := UnscopedDb().Table("folders").
		Select("folders.path, folders.root, folders.folder_uid, folders.folder_title, folders.folder_country, folders.folder_year, folders.folder_month, COUNT(photos.id) AS photo_count").
		Joins("JOIN photos ON photos.photo_path = folders.path AND photos.deleted_at IS NULL AND photos.photo_quality >= 3").
		Group("folders.path, folders.root, folders.folder_uid, folders.folder_title, folders.folder_country, folders.folder_year, folders.folder_month").
		Having("photo_count >= ?", threshold)

	if err := db.Scan(&folders).Error; err != nil {
		return folders, err
	}

	return folders, nil
}
