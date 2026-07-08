package query

import (
	"errors"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
)

// SetDownloadFileID updates the local file id for remote downloads.
func SetDownloadFileID(filename string, fileId uint) error {
	if len(filename) == 0 {
		return errors.New("sync: cannot update, filename empty")
	}

	filename = clean.SlashPath(filename)

	filename = "/" + filename

	result := Db().Model(entity.FileSync{}).
		Where("remote_name = ? AND status = ? AND file_id = 0", filename, entity.FileSyncDownloaded).
		Update("file_id", fileId)

	return result.Error
}
