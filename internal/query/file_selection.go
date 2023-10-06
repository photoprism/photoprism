package query

import (
	"errors"
	"fmt"

	"github.com/photoprism/photoprism/pkg/fs"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/media"
)

const MegaByte = 1024 * 1024

// FileSelection represents a selection filter to include/exclude certain files.
type FileSelection struct {
	MaxSize   int
	Media     []string
	OmitMedia []string
	Types     []string
	OmitTypes []string
	Primary   bool
	Originals bool
	Hidden    bool
	Private   bool
	Archived  bool
}

// DownloadSelection selects files to download.
func DownloadSelection(mediaRaw, mediaSidecar, originals bool) FileSelection {
	omitMedia := make([]string, 0, 2)

	if !mediaRaw {
		omitMedia = append(omitMedia, media.Raw.String())
	}

	if !mediaSidecar {
		omitMedia = append(omitMedia, media.Sidecar.String())
	}

	return FileSelection{
		OmitMedia: omitMedia,
		Originals: originals,
		Private:   true,
		Archived:  true,
		Hidden:    true,
	}
}

// ShareSelection selects files to share, for example for upload via WebDAV.
func ShareSelection(originals bool) FileSelection {
	var omitMedia []string
	var omitTypes []string

	if !originals {
		omitMedia = []string{
			media.Unknown.String(),
			media.Raw.String(),
			media.Sidecar.String(),
		}

		omitTypes = []string{
			fs.ImagePNG.String(),
			fs.ImageWebP.String(),
			fs.ImageTIFF.String(),
			fs.ImageAVIF.String(),
			fs.ImageHEIC.String(),
			fs.ImageBMP.String(),
			fs.ImageGIF.String(),
		}
	}

	return FileSelection{
		Originals: originals,
		OmitMedia: omitMedia,
		OmitTypes: omitTypes,
		Hidden:    false,
		Private:   false,
		Archived:  false,
		MaxSize:   1024 * MegaByte,
	}
}

// SelectedFiles finds files based on the given selection form, e.g. for downloading or sharing.
func SelectedFiles(f form.Selection, o FileSelection) (results entity.Files, err error) {
	if f.Empty() {
		return results, errors.New("no items selected")
	}

	// Resolve photos in smart albums.
	if photoIds, err := AlbumsPhotoUIDs(f.Albums, false, o.Private); err != nil {
		log.Warnf("query: %s", err.Error())
	} else if len(photoIds) > 0 {
		f.Photos = append(f.Photos, photoIds...)
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

	// Search condition.
	where := fmt.Sprintf(`photos.photo_uid IN (?) 
		OR photos.place_id IN (?) 
		OR photos.photo_uid IN (SELECT photo_uid FROM files WHERE file_uid IN (?))
		OR photos.photo_path IN (
			SELECT a.path FROM folders a WHERE a.folder_uid IN (?) UNION
			SELECT b.path FROM folders a JOIN folders b ON b.path LIKE %s WHERE a.folder_uid IN (?))
		OR photos.photo_uid IN (SELECT photo_uid FROM photos_albums WHERE hidden = 0 AND album_uid IN (?))
		OR files.file_uid IN (SELECT file_uid FROM %s m WHERE m.subj_uid IN (?))
		OR photos.id IN (SELECT pl.photo_id FROM photos_labels pl JOIN labels l ON pl.label_id = l.id AND pl.uncertainty < 100 AND l.deleted_at IS NULL WHERE l.label_uid IN (?))
		OR photos.id IN (SELECT pl.photo_id FROM photos_labels pl JOIN categories c ON c.label_id = pl.label_id AND pl.uncertainty < 100 JOIN labels lc ON lc.id = c.category_id AND lc.deleted_at IS NULL WHERE lc.label_uid IN (?))`,
		concat, entity.Marker{}.TableName())

	// Build search query.
	s := UnscopedDb().Table("files").
		Select("files.*").
		Joins("JOIN photos ON photos.id = files.photo_id").
		Where("files.file_missing = 0 AND files.file_name <> '' AND files.file_hash <> ''").
		Where(where, f.Photos, f.Places, f.Files, f.Files, f.Files, f.Albums, f.Subjects, f.Labels, f.Labels).
		Group("files.id")

	// File size limit?
	if o.MaxSize > 0 {
		s = s.Where("files.file_size < ?", o.MaxSize)
	}

	// Specific media types only?
	if len(o.Media) > 0 {
		s = s.Where("files.media_type IN (?)", o.Media)
	}

	// Exclude media types?
	if len(o.OmitMedia) > 0 {
		s = s.Where("files.media_type NOT IN (?)", o.OmitMedia)
	}

	// Specific file types only?
	if len(o.Types) > 0 {
		s = s.Where("files.file_type IN (?)", o.Types)
	}

	// Exclude file types?
	if len(o.OmitTypes) > 0 {
		s = s.Where("files.file_type NOT IN (?)", o.OmitTypes)
	}

	// Previews files only?
	if o.Primary {
		s = s.Where("files.file_primary = 1")
	}

	// Files in originals only?
	if o.Originals {
		s = s.Where("files.file_root = '/'")
	}

	// Exclude private?
	if !o.Private {
		s = s.Where("photos.photo_private <> 1")
	}

	// Exclude hidden photos?
	if !o.Hidden {
		s = s.Where("photos.photo_quality > -1")
	}

	// Exclude archived photos?
	if !o.Archived {
		s = s.Where("photos.deleted_at IS NULL")
	}

	// Find and return.
	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
