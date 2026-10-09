package batch

import (
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestPhotoLabelWriteRetry checks the retry wiring of both batch assignment operations.
func TestPhotoLabelWriteRetry(t *testing.T) {
	for _, operation := range []string{"Update", "Delete"} {
		t.Run(operation, func(t *testing.T) {
			photo := entity.NewPhoto(false)
			require.NoError(t, photo.Save())
			label := entity.NewLabel("Batch Retry "+rnd.GenerateUID('l'), 0)
			require.NoError(t, label.Create())
			link := entity.NewPhotoLabel(photo.ID, label.ID, 40, entity.SrcImage)
			require.NoError(t, link.Create())
			t.Cleanup(func() {
				entity.WaitForAsyncJobs()
				entity.UnscopedDb().Delete(&entity.PhotoLabel{}, "photo_id = ?", photo.ID)
				entity.UnscopedDb().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
				entity.UnscopedDb().Delete(&photo)
				entity.UnscopedDb().Delete(label)
			})
			attempts := 0
			callback := func(db *gorm.DB) {
				if db.Statement.Table != (entity.PhotoLabel{}).TableName() {
					return
				}
				attempts++
				if attempts == 1 {
					_ = db.Statement.AddError(&mysql.MySQLError{Number: 1213, Message: "batch write control"})
				}
			}
			if operation == "Update" {
				entity.Db().Callback().Update().Before("gorm:begin_transaction").Register("test:batch-retry", callback)
				t.Cleanup(func() { entity.Db().Callback().Update().Remove("test:batch-retry") })
				link.LabelSrc, link.Uncertainty = entity.SrcBatch, 0
				require.NoError(t, updatePhotoLabel(link, "test update"))
				var stored entity.PhotoLabel
				require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&stored).Error)
				assert.Zero(t, stored.Uncertainty)
				assert.Equal(t, entity.SrcBatch, stored.LabelSrc)
			} else {
				require.NoError(t, entity.Db().Callback().Delete().Before("gorm:begin_transaction").Register("test:batch-retry", callback))
				t.Cleanup(func() { require.NoError(t, entity.Db().Callback().Delete().Remove("test:batch-retry")) })
				require.NoError(t, deletePhotoLabel(link))
				var count int64
				require.NoError(t, entity.Db().Model(&entity.PhotoLabel{}).Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).Count(&count).Error)
				assert.Zero(t, count)
			}
			assert.Equal(t, 2, attempts)
		})
	}
	t.Run("Nil", func(t *testing.T) {
		require.Error(t, updatePhotoLabel(nil, "test"))
		require.Error(t, deletePhotoLabel(nil))
	})
}
