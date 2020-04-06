package query

import (
	"errors"
	"os"

	"github.com/photoprism/photoprism/internal/entity"
)

// FileSyncs returns up to 100 file syncs for a given account id and status.
func (q *Query) FileSyncs(accountId uint, status string) (result []entity.FileSync, err error) {
	s := q.db.Where(&entity.FileSync{})

	if accountId > 0 {
		s = s.Where("account_id = ?", accountId)
	}

	if status != "" {
		s = s.Where("status = ?", status)
	}

	s = s.Order("remote_name ASC")
	s = s.Limit(1000).Offset(0)

	s = s.Preload("File")

	if err := s.Find(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}

// SetDownloadFileID updates the local file id for remote downloads.
func (q *Query) SetDownloadFileID(filename string, fileId uint) error {
	if len(filename) == 0 {
		return errors.New("sync: can't update, filename empty")
	}

	// TODO: Might break on Windows
	if filename[0] != os.PathSeparator {
		filename = string(os.PathSeparator) + filename
	}

	result := q.db.Model(entity.FileSync{}).
		Where("remote_name = ? AND status = ? AND file_id = 0", filename, entity.FileSyncDownloaded).
		Update("file_id", fileId)

	return result.Error
}
