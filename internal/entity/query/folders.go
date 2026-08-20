package query

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
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

		if err = newFolder.Create(); err == nil {
			folders[i] = newFolder
		} else if folder := entity.FindFolder(rootName, filepath.Join(path, dir)); folder != nil {
			folders[i] = *folder
		} else {
			log.Errorf("folders: %s (create folder)", err)
		}
	}

	return folders, nil
}

// FoldersByRoot returns all folders for the specified Root, optionally including deleted.
func FoldersByRoot(root string, includeDeleted bool) (results entity.Folders, err error) {
	if includeDeleted {
		err = UnscopedDb().Where("root = ?", root).Find(&results).Error
	} else {
		err = Db().Where("root = ?", root).Find(&results).Error
	}
	return results, err
}

// FolderCoverByUID returns a folder cover file based on the uid.
func FolderCoverByUID(uid string) (file entity.File, err error) {
	if rnd.InvalidUID(uid, entity.FolderUID) {
		return file, fmt.Errorf("invalid folder uid")
	}

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
func UpdateFolderDates() (updated int, err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	switch DbDialect() {
	case dsn.DriverMySQL:
		result := UnscopedDb().Exec(`UPDATE folders
		INNER JOIN
			(SELECT photo_path, MAX(taken_at_local) AS taken_max
			FROM photos WHERE taken_src = 'meta' AND photos.photo_quality >= 3 AND photos.deleted_at IS NULL
			GROUP BY photo_path) AS p ON folders.path = p.photo_path
		SET folders.folder_year = YEAR(taken_max), folders.folder_month = MONTH(taken_max), folders.folder_day = DAY(taken_max)
		WHERE p.taken_max IS NOT NULL AND root = ?
		AND (folder_year = 0 OR folder_month = 0 OR folder_day = 0
			OR DATE(p.taken_max) <> COALESCE(STR_TO_DATE(CONCAT(folder_year, '-', folder_month,'-', folder_day), '%Y-%c-%e'), DATE('1000-01-01'))
		)`, entity.RootOriginals)
		return int(result.RowsAffected), result.Error
	case dsn.DriverSQLite3:
		// SQLite has potential locking issues if the update is done on all folders at once.
		var folders entity.Folders
		// Only update Original's folders.
		if folders, err = FoldersByRoot(entity.RootOriginals, true); err != nil {
			log.Errorf("folders: get folders (%v)", err)
			return 0, err
		}

		var photos map[string]time.Time
		if photos, err = photoPathMaxDates(); err != nil {
			return updated, err
		}
		pathCount := 0
		tx := UnscopedDb().Begin()
		if tx.Error != nil {
			log.Errorf("folders: begin transaction failed (%v)", tx.Error)
			return updated, tx.Error
		}
		for _, folder := range folders {
			if pathCount == 5000 {
				log.Debugf("folders: committing")
				if err = tx.Commit().Error; err != nil {
					log.Errorf("folders: commit partial changes failed (%v)", err)
					return updated, err
				}
				tx = UnscopedDb().Begin()
				if tx.Error != nil {
					log.Errorf("folders: begin transaction failed (%v)", tx.Error)
					return updated, tx.Error
				}
				pathCount = 0
			}
			takenMax, ok := photos[folder.Path]
			if ok {
				var folderDate time.Time
				if folderDate, err = time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-%02d", folder.FolderYear, folder.FolderMonth, folder.FolderDay)); err != nil {
					// Date wasn't valid, so set to 1000-01-01
					folderDate = time.Date(1000, time.January, 1, 0, 0, 0, 0, time.UTC)
				}
				if takenMax != folderDate {
					result := tx.Exec(`UPDATE folders SET folder_year = ?, folder_month = ?, folder_day = ? WHERE folder_uid = ?`, takenMax.Year(), takenMax.Month(), takenMax.Day(), folder.FolderUID)
					if err = result.Error; err != nil {
						log.Errorf("folders: set folder dates on %s (%v)", clean.Log(folder.FolderTitle), err)
						if errRB := tx.Rollback(); errRB != nil {
							log.Errorf("folders: rollback changes failed (%v)", errRB)
						}
						return updated, err
					} else {
						updated += int(result.RowsAffected)
					}
				}
			}
		}
		return updated, tx.Commit().Error
	default:
		return updated, err
	}
}
