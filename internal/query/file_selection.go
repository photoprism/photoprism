package query

import (
	"errors"
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// FileSelection represents a selection filter to include/exclude certain files.
type FileSelection struct {
	Video         bool
	Sidecar       bool
	PrimaryOnly   bool
	OriginalsOnly bool
	SizeLimit     int
	Include       []string
	Exclude       []string
}

// FileSelectionAll returns options that include videos and sidecar files.
func FileSelectionAll() FileSelection {
	return FileSelection{
		Video:         true,
		Sidecar:       true,
		PrimaryOnly:   false,
		OriginalsOnly: false,
		SizeLimit:     0,
		Include:       []string{},
		Exclude:       []string{},
	}
}

// SelectedFiles finds files based on the given selection form, e.g. for downloading or sharing.
func SelectedFiles(f form.Selection, o FileSelection) (results entity.Files, err error) {
	if f.Empty() {
		return results, errors.New("no items selected")
	}

	var concat string

	switch DbDialect() {
	case MySQL:
		concat = "CONCAT(a.path, '/%')"
	case SQLite3:
		concat = "a.path || '/%'"
	default:
		return results, fmt.Errorf("unknown sql dialect: %s", DbDialect())
	}

	where := fmt.Sprintf(`photos.photo_uid IN (?) 
		OR photos.place_id IN (?) 
		OR photos.photo_uid IN (SELECT photo_uid FROM files WHERE file_uid IN (?))
		OR photos.photo_path IN (
			SELECT a.path FROM folders a WHERE a.folder_uid IN (?) UNION
			SELECT b.path FROM folders a JOIN folders b ON b.path LIKE %s WHERE a.folder_uid IN (?))
		OR photos.photo_uid IN (SELECT photo_uid FROM photos_albums WHERE hidden = 0 AND album_uid IN (?))
		OR files.file_uid IN (SELECT file_uid FROM %s m WHERE m.subj_uid IN (?))
		OR photos.id IN (SELECT pl.photo_id FROM photos_labels pl JOIN labels l ON pl.label_id = l.id AND l.deleted_at IS NULL WHERE l.label_uid IN (?))
		OR photos.id IN (SELECT pl.photo_id FROM photos_labels pl JOIN categories c ON c.label_id = pl.label_id JOIN labels lc ON lc.id = c.category_id AND lc.deleted_at IS NULL WHERE lc.label_uid IN (?))`,
		concat, entity.Marker{}.TableName())

	s := UnscopedDb().Table("files").
		Select("files.*").
		Joins("JOIN photos ON photos.id = files.photo_id").
		Where("photos.deleted_at IS NULL").
		Where("files.file_missing = 0").
		Where(where, f.Photos, f.Places, f.Files, f.Files, f.Files, f.Albums, f.Subjects, f.Labels, f.Labels).
		Group("files.id")

	if o.OriginalsOnly {
		s = s.Where("file_root = '/'")
	}

	if o.PrimaryOnly {
		s = s.Where("file_primary = 1")
	}

	if !o.Sidecar {
		s = s.Where("file_sidecar = 0")
	}

	if !o.Video {
		s = s.Where("file_video = 0")
	}

	if o.SizeLimit > 0 {
		s = s.Where("file_size < ?", o.SizeLimit)
	}

	if len(o.Include) > 0 {
		s = s.Where("file_type IN (?)", o.Include)
	}

	if len(o.Exclude) > 0 {
		s = s.Where("file_type NOT IN (?)", o.Exclude)
	}

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
