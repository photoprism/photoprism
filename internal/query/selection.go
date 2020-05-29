package query

import (
	"errors"

	"github.com/photoprism/photoprism/internal/form"
)

// PhotoSelection queries all selected photos.
func PhotoSelection(f form.Selection) (results Photos, err error) {
	if f.Empty() {
		return results, errors.New("no items selected")
	}

	// Db().LogMode(true)

	s := UnscopedDb().Table("photos").
		Select("photos.*").
		Where(`photos.photo_uid IN (?) 
		OR photos.place_id IN (?) 
		OR photos.photo_uid IN (SELECT photo_uid FROM files WHERE file_uid IN (?))
		OR photos.photo_path IN (
			SELECT a.path FROM folders a WHERE a.folder_uid IN (?) UNION
			SELECT b.path FROM folders a JOIN folders b ON b.path LIKE CONCAT(a.path, '/%') WHERE a.folder_uid IN (?))
		OR photos.photo_uid IN (SELECT photo_uid FROM photos_albums WHERE album_uid IN (?))
		OR photos.id IN (SELECT pl.photo_id FROM photos_labels pl JOIN labels l ON pl.label_id = l.id AND l.deleted_at IS NULL WHERE l.label_uid IN (?))
		OR photos.id IN (SELECT pl.photo_id FROM photos_labels pl JOIN categories c ON c.label_id = pl.label_id JOIN labels lc ON lc.id = c.category_id AND lc.deleted_at IS NULL WHERE lc.label_uid IN (?))`,
			f.Photos, f.Places, f.Files, f.Files, f.Files, f.Albums, f.Labels, f.Labels)

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}

// FileSelection queries all selected files e.g. for downloading.
func FileSelection(f form.Selection) (results Files, err error) {
	if f.Empty() {
		return results, errors.New("no items selected")
	}

	// Db().LogMode(true)

	s := UnscopedDb().Table("files").
		Select("files.*").
		Joins("JOIN photos ON photos.id = files.photo_id").
		Where("photos.deleted_at IS NULL").
		Where("files.file_missing = 0").
		Where(`photos.photo_uid IN (?) 
		OR photos.place_id IN (?) 
		OR photos.photo_uid IN (SELECT photo_uid FROM files WHERE file_uid IN (?))
		OR photos.photo_path IN (
			SELECT a.path FROM folders a WHERE a.folder_uid IN (?) UNION
			SELECT b.path FROM folders a JOIN folders b ON b.path LIKE CONCAT(a.path, '/%') WHERE a.folder_uid IN (?))
		OR photos.photo_uid IN (SELECT photo_uid FROM photos_albums WHERE album_uid IN (?))
		OR photos.id IN (SELECT pl.photo_id FROM photos_labels pl JOIN labels l ON pl.label_id = l.id AND l.deleted_at IS NULL WHERE l.label_uid IN (?))
		OR photos.id IN (SELECT pl.photo_id FROM photos_labels pl JOIN categories c ON c.label_id = pl.label_id JOIN labels lc ON lc.id = c.category_id AND lc.deleted_at IS NULL WHERE lc.label_uid IN (?))`,
			f.Photos, f.Places, f.Files, f.Files, f.Files, f.Albums, f.Labels, f.Labels).
		Group("files.id")

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
