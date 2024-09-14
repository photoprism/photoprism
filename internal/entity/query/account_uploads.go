package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

// AccountUploads a list of files for uploading to a remote account.
func AccountUploads(a entity.Service, limit int) (results entity.Files, err error) {
	s := Db().Where("files.file_missing = 0").
		Where("files.id NOT IN (SELECT file_id FROM files_sync WHERE file_id > 0 AND service_id = ?)", a.ID)

	if !a.SyncRaw {
		s = s.Where("files.file_type <> ? OR files.file_type IS NULL", fs.ImageRaw)
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
