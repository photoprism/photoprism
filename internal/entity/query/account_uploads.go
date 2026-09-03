package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/media"
)

// AccountUploads a list of files for uploading to a remote account.
func AccountUploads(a entity.Service, limit int) (results entity.Files, err error) {
	s := Db().Where("files.file_missing = FALSE").
		Where("files.id NOT IN (SELECT file_id FROM files_sync WHERE file_id > 0 AND service_id = ?)", a.ID)

	if !a.SyncRaw {
		// Hold back raw images and standalone videos, gating on the media type like the
		// download direction. Live photos and animations stored as mp4/mov classify as
		// "live"/"animated" and still upload. The file type is matched as well because
		// media_type is not backfilled on rows indexed before it existed.
		s = s.Where("(files.file_type NOT IN (?) OR files.file_type IS NULL) AND (files.media_type NOT IN (?) OR files.media_type IS NULL)",
			media.FileTypeStrings(media.Raw), []string{media.Raw.String(), media.Video.String()})
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
