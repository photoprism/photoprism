package query

import (
	"fmt"
	"path"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// FilesByPath returns a slice of files in a given originals folder.
func FilesByPath(limit, offset int, rootName, pathName string, public bool) (files entity.Files, err error) {
	if strings.HasPrefix(pathName, "/") {
		pathName = pathName[1:]
	}

	stmt := Db().
		Table("files").Select("files.*").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.deleted_at IS NULL").
		Where("files.file_missing = 0 AND files.file_root = ?", rootName).
		Where("photos.photo_path = ?", pathName)

	if public {
		stmt = stmt.Where("photos.photo_private = 0")
	}

	err = stmt.Order("files.file_name").
		Limit(limit).Offset(offset).
		Find(&files).Error

	return files, err
}

// Files returns not-missing and not-deleted file entities in the range of limit and offset sorted by id.
func Files(limit, offset int, pathName string, includeMissing bool) (files entity.Files, err error) {
	if strings.HasPrefix(pathName, "/") {
		pathName = pathName[1:]
	}

	stmt := Db()

	if !includeMissing {
		stmt = stmt.Where("file_missing = 0")
	}

	if pathName != "" {
		stmt = stmt.Where("files.file_name LIKE ?", pathName+"/%")
	}

	err = stmt.Order("id").Limit(limit).Offset(offset).Find(&files).Error

	return files, err
}

// FilesByUID finds files for the given UIDs.
func FilesByUID(u []string, limit int, offset int) (files entity.Files, err error) {
	if err = Db().Where("(photo_uid IN (?) AND file_primary = 1) OR file_uid IN (?)", u, u).Preload("Photo").Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		return files, err
	}

	return files, nil
}

// FileByPhotoUID finds a file for the given photo UID.
func FileByPhotoUID(photoUID string) (*entity.File, error) {
	f := entity.File{}

	if photoUID == "" {
		return &f, fmt.Errorf("photo uid required")
	}

	err := Db().Where("photo_uid = ? AND file_primary = 1", photoUID).Preload("Photo").First(&f).Error

	return &f, err
}

// VideoByPhotoUID finds a video for the given photo UID.
func VideoByPhotoUID(photoUID string) (*entity.File, error) {
	f := entity.File{}

	if photoUID == "" {
		return &f, fmt.Errorf("photo uid required")
	}

	err := Db().Where("photo_uid = ? AND file_missing = 0", photoUID).
		Where("file_video = 1 OR file_duration > 0 OR file_frames > 0 OR file_type = ?", fs.ImageGIF).
		Order("file_error ASC, file_video DESC, file_duration DESC, file_frames DESC").
		Preload("Photo").First(&f).Error

	return &f, err
}

// FileByUID finds a file entity for the given UID.
func FileByUID(fileUID string) (*entity.File, error) {
	f := entity.File{}

	if fileUID == "" {
		return &f, fmt.Errorf("file uid required")
	}

	err := Db().Where("file_uid = ?", fileUID).Preload("Photo").First(&f).Error

	return &f, err
}

// FileByHash finds a file with a given hash string.
func FileByHash(fileHash string) (*entity.File, error) {
	f := entity.File{}

	if fileHash == "" {
		return &f, fmt.Errorf("file hash required")
	}

	err := Db().Where("file_hash = ?", fileHash).Preload("Photo").First(&f).Error

	return &f, err
}

// RenameFile renames an indexed file.
func RenameFile(srcRoot, srcName, destRoot, destName string) error {
	if srcRoot == "" || srcName == "" || destRoot == "" || destName == "" {
		return fmt.Errorf("cannot rename %s/%s to %s/%s", srcRoot, srcName, destRoot, destName)
	}

	return Db().Exec("UPDATE files SET file_root = ?, file_name = ?, file_missing = 0, deleted_at = NULL WHERE file_root = ? AND file_name = ?", destRoot, destName, srcRoot, srcName).Error
}

// SetPhotoPrimary sets a new primary image file for a photo.
func SetPhotoPrimary(photoUID, fileUID string) (err error) {
	if photoUID == "" {
		return fmt.Errorf("photo uid is missing")
	}

	var files []string

	if fileUID != "" {
		// Do nothing.
	} else if err = Db().Model(entity.File{}).
		Where("photo_uid = ? AND file_missing = 0 AND file_type IN (?)", photoUID, media.PreviewExpr).
		Order("file_width DESC, file_hdr DESC").Limit(1).Pluck("file_uid", &files).Error; err != nil {
		return err
	} else if len(files) == 0 {
		return fmt.Errorf("cannot find primary file for %s", photoUID)
	} else {
		fileUID = files[0]
	}

	if fileUID == "" {
		return fmt.Errorf("file uid is missing")
	}

	if err = Db().Model(entity.File{}).
		Where("photo_uid = ? AND file_uid <> ?", photoUID, fileUID).
		UpdateColumn("file_primary", 0).Error; err != nil {
		return err
	} else if err = Db().
		Model(entity.File{}).Where("photo_uid = ? AND file_uid = ?", photoUID, fileUID).
		UpdateColumn("file_primary", 1).Error; err != nil {
		return err
	} else {
		entity.File{PhotoUID: photoUID}.RegenerateIndex()
	}

	return nil
}

// SetFileError updates the file error column.
func SetFileError(fileUID, errorString string) {
	if err := Db().Model(entity.File{}).Where("file_uid = ?", fileUID).UpdateColumn("file_error", errorString).Error; err != nil {
		log.Errorf("files: %s (set error)", err.Error())
	}
}

type FileMap map[string]int64

// IndexedFiles returns a map of already indexed files with their mod time unix timestamp as value.
func IndexedFiles() (result FileMap, err error) {
	result = make(FileMap)

	type File struct {
		FileRoot string
		FileName string
		ModTime  int64
	}

	// Query known duplicates.
	var duplicates []File

	if err := UnscopedDb().Raw("SELECT file_root, file_name, mod_time FROM duplicates").Scan(&duplicates).Error; err != nil {
		return result, err
	}

	for _, row := range duplicates {
		result[path.Join(row.FileRoot, row.FileName)] = row.ModTime
	}

	// Query indexed files.
	var files []File

	if err := UnscopedDb().Raw("SELECT file_root, file_name, mod_time FROM files WHERE file_missing = 0 AND deleted_at IS NULL").Scan(&files).Error; err != nil {
		return result, err
	}

	for _, row := range files {
		result[path.Join(row.FileRoot, row.FileName)] = row.ModTime
	}

	return result, err
}

// OrphanFiles finds files without a photo.
func OrphanFiles() (files entity.Files, err error) {
	err = UnscopedDb().
		Raw(`SELECT * FROM files WHERE photo_id NOT IN (SELECT id FROM photos)`).
		Find(&files).Error

	return files, err
}
