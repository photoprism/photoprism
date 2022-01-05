package query

import (
	"errors"
	"os"

	"github.com/photoprism/photoprism/internal/entity"
)

// SetDownloadFileID updates the local file id for remote downloads.
func SetDownloadFileID(filename string, fileId uint) error {
	if len(filename) == 0 {
		return errors.New("sync: cannot update, filename empty")
	}

	// TODO: Might break on Windows
	if filename[0] != os.PathSeparator {
		filename = string(os.PathSeparator) + filename
	}

	result := Db().Model(entity.FileSync{}).
		Where("remote_name = ? AND status = ? AND file_id = 0", filename, entity.FileSyncDownloaded).
		Update("file_id", fileId)

	return result.Error
}
