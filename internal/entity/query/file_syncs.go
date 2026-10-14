package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// FileSyncs returns a list of FileSync entities for a given account and status, with the fewest errors first.
func FileSyncs(accountId uint, status string, limit int) (result []entity.FileSync, err error) {
	s := Db().Where(&entity.FileSync{})

	if accountId > 0 {
		s = s.Where("service_id = ?", accountId)
	}

	if status != "" {
		s = s.Where("status = ?", status)
	}

	// Rows without an error count sort with those not tried yet on every database.
	s = s.Order("COALESCE(errors, 0) ASC, remote_name ASC")

	if limit > 0 {
		s = s.Limit(limit).Offset(0)
	}

	s = s.Preload("File")

	if err := s.Find(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}

// CountQueuedFileSyncs returns the number of files of an account that are queued for download, of those that
// were not tried yet, and of those that failed.
func CountQueuedFileSyncs(accountId uint) (queued, untried, failed int64, err error) {
	var counts struct {
		Queued  int64
		Untried int64
		Failed  int64
	}

	err = Db().Model(&entity.FileSync{}).
		Select("COUNT(*) AS queued, "+
			"COALESCE(SUM(CASE WHEN COALESCE(errors, 0) = 0 THEN 1 ELSE 0 END), 0) AS untried, "+
			"COALESCE(SUM(CASE WHEN errors > 0 THEN 1 ELSE 0 END), 0) AS failed").
		Where("service_id = ? AND status = ?", accountId, entity.FileSyncNew).
		Scan(&counts).Error

	return counts.Queued, counts.Untried, counts.Failed, err
}
