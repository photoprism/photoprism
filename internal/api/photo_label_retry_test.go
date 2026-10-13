package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestUpdatePhotoLabel_RetryDeadlock checks that each write of a label update is retried after a deadlock.
func TestUpdatePhotoLabel_RetryDeadlock(t *testing.T) {
	photo := entity.NewPhoto(false)
	require.NoError(t, photo.Save())
	label := entity.NewLabel("Retry "+rnd.GenerateUID('l'), 0)
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

	tables := []string{entity.PhotoLabel{}.TableName(), entity.Label{}.TableName(), entity.Photo{}.TableName()}
	attempts := map[string]int{}
	for _, table := range tables {
		attempts[table] = 0
	}
	_ = entity.Db().Callback().Update().Before("gorm:begin_transaction").Register("test:label-retry", func(tx *gorm.DB) {
		if _, ok := attempts[tx.Statement.Table]; !ok {
			return
		} else if attempts[tx.Statement.Table]++; attempts[tx.Statement.Table] == 1 {
			_ = tx.AddError(&mysql.MySQLError{Number: 1213, Message: "deadlock control"})
		}
	})
	t.Cleanup(func() { _ = entity.Db().Callback().Update().Remove("test:label-retry") })

	app, router, _ := NewApiTest()
	UpdatePhotoLabel(router)
	r := PerformRequestWithBody(app, "PUT", fmt.Sprintf("/api/v1/photos/%s/label/%d", photo.PhotoUID, label.ID),
		`{"Uncertainty": 0, "Label": {"Name": "Retry Rename Control"}}`)
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())
	for _, table := range tables {
		assert.Equal(t, 2, attempts[table], table)
	}

	var stored entity.PhotoLabel
	require.NoError(t, entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&stored).Error)
	assert.Zero(t, stored.Uncertainty)
	assert.Equal(t, entity.SrcManual, stored.LabelSrc)
	var renamed entity.Label
	require.NoError(t, entity.Db().First(&renamed, label.ID).Error)
	assert.Equal(t, "Retry Rename Control", renamed.LabelName)
}
