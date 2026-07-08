package query

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestSetDownloadFileID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		err := SetDownloadFileID("exampleFileName.jpg", 1000000)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("FilenameEmpty", func(t *testing.T) {
		err := SetDownloadFileID("", 1000000)
		if err == nil {
			t.Fatal()
		}
		assert.Equal(t, "sync: cannot update, filename empty", err.Error())
	})
	t.Run("NormalizesBackslashes", func(t *testing.T) {
		serviceID := uint(1000001)
		remoteName := "/sync-normalize-" + time.Now().UTC().Format("20060102150405.000000000") + "/file.jpg"

		entry := entity.FileSync{
			ServiceID:  serviceID,
			RemoteName: remoteName,
			Status:     entity.FileSyncDownloaded,
			FileID:     0,
		}

		if err := entity.Db().Create(&entry).Error; err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			_ = entity.UnscopedDb().
				Where("service_id = ? AND remote_name = ?", serviceID, remoteName).
				Delete(&entity.FileSync{}).Error
		})

		// Keep lookup stable: use the path of the created row with backslashes.
		incoming := strings.ReplaceAll(remoteName[1:], "/", "\\")

		if err := SetDownloadFileID(incoming, 1000000); err != nil {
			t.Fatal(err)
		}

		var updated entity.FileSync
		if err := entity.Db().Where("service_id = ? AND remote_name = ?", serviceID, remoteName).First(&updated).Error; err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint(1000000), updated.FileID)
	})
}
