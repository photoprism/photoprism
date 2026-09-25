package query

import (
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

// FileShares returns up to 100 file shares for a given account id and status.
func FileShares(accountId uint, status string) (result []entity.FileShare, err error) {
	s := Db().Where(&entity.FileShare{})

	if accountId > 0 {
		s = s.Where("service_id = ?", accountId)
	}

	if status != "" {
		s = s.Where("status = ?", status)
	}

	s = s.Order("created_at ASC")
	s = s.Limit(100).Offset(0)

	s = s.Preload("File")

	if err := s.Find(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}

// QueuedFileShares returns up to 100 queued file shares for an account, omitting YAML sidecar files unless it syncs them.
func QueuedFileShares(account entity.Service) (result []entity.FileShare, err error) {
	s := Db().Model(&entity.FileShare{}).Select("files_share.*").
		Where("files_share.service_id = ? AND files_share.status = ?", account.ID, entity.FileShareNew)

	if !account.SyncYaml {
		s = s.Joins("LEFT JOIN files ON files.id = files_share.file_id").
			Where("(files.file_type <> ? OR files.file_type IS NULL)", fs.SidecarYaml.String())
	}

	s = s.Order("files_share.created_at ASC")
	s = s.Limit(100).Offset(0)

	s = s.Preload("File")

	if err = s.Find(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}

// ExpiredFileShares returns up to 100 expired file shares for a given account.
func ExpiredFileShares(account entity.Service) (result []entity.FileShare, err error) {
	if account.ShareExpires <= 0 {
		return result, nil
	}

	s := Db().Where(&entity.FileShare{})

	exp := time.Now().Add(time.Duration(-1*account.ShareExpires) * time.Second)

	s = s.Where("service_id = ?", account.ID)
	s = s.Where("status = ?", entity.FileShareShared)
	s = s.Where("updated_at < ?", exp)

	s = s.Order("updated_at ASC")
	s = s.Limit(100).Offset(0)

	s = s.Preload("File")

	if err := s.Find(&result).Error; err != nil {
		return result, err
	}

	return result, nil
}
