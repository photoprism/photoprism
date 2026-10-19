package query

import (
	"fmt"

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

// QueuedFileSyncCounts counts the files of an account that are queued for download.
type QueuedFileSyncCounts struct {
	Queued    int64 // Files queued for download.
	Untried   int64 // Queued files that were not tried yet.
	Failed    int64 // Queued files that failed at least once.
	Retryable int64 // Queued files that may still be tried with the retry limit.
	Errors    int64 // Errors of the queued files in total.
}

// CountQueuedFileSyncs counts the files of an account that are queued for download, where a positive retry limit
// decides which of them may still be tried.
func CountQueuedFileSyncs(accountId uint, retryLimit int) (counts QueuedFileSyncCounts, err error) {
	retryable := "1"

	if retryLimit > 0 {
		retryable = fmt.Sprintf("CASE WHEN COALESCE(errors, 0) <= %d THEN 1 ELSE 0 END", retryLimit)
	}

	err = Db().Model(&entity.FileSync{}).
		Select("COUNT(*) AS queued, "+
			"COALESCE(SUM(CASE WHEN COALESCE(errors, 0) = 0 THEN 1 ELSE 0 END), 0) AS untried, "+
			"COALESCE(SUM(CASE WHEN errors > 0 THEN 1 ELSE 0 END), 0) AS failed, "+
			"COALESCE(SUM("+retryable+"), 0) AS retryable, "+
			"COALESCE(SUM(COALESCE(errors, 0)), 0) AS errors").
		Where("service_id = ? AND status = ?", accountId, entity.FileSyncNew).
		Scan(&counts).Error

	return counts, err
}
