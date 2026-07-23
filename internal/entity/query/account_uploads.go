package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// AccountUploads a list of files for uploading to a remote account.
func AccountUploads(a entity.Service, limit int) (results entity.Files, err error) {
	s := Db().Where("files.file_missing = 0").
		Where("files.id NOT IN (SELECT file_id FROM files_sync WHERE file_id > 0 AND service_id = ?)", a.ID)

	if !a.SyncRaw {
		// Hold back raw images (by file type) and videos (by media type). Live photos and
		// animations may be stored as mp4/mov but classify as "live"/"animated", so they
		// are still uploaded, matching the media-class gating used for downloads.
		s = s.Where("(files.file_type <> ? OR files.file_type IS NULL) AND (files.media_type <> ? OR files.media_type IS NULL)",
			fs.ImageRaw, media.Video.String())
	}

	s = s.Order("files.file_name ASC")

	if limit > 0 {
		s = s.Limit(limit).Offset(0)
	}

	if result := s.Find(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
