package entity

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestPhotoResetCaption(t *testing.T) {
	photo := createTestPhoto(t)

	require.NoError(t, Db().Model(&photo).Updates(Values{
		"PhotoCaption": "Generated caption",
		"CaptionSrc":   SrcOllama,
	}).Error)

	label := NewLabel(fmt.Sprintf("reset-caption-%s", rnd.GenerateUID(LabelUID)), 10)
	require.NoError(t, label.Create())

	t.Cleanup(func() {
		_ = Db().Delete(&PhotoLabel{}, "photo_id = ?", photo.ID).Error
		_ = Db().Delete(label).Error
	})

	require.NotNil(t, FirstOrCreatePhotoLabel(NewPhotoLabel(photo.ID, label.ID, 10, SrcCaption)))

	changed, err := photo.ResetCaption(SrcOllama)
	require.NoError(t, err)
	assert.True(t, changed)

	var refreshed Photo
	require.NoError(t, Db().First(&refreshed, photo.ID).Error)
	assert.Empty(t, refreshed.PhotoCaption)
	assert.Empty(t, refreshed.CaptionSrc)

	var count int
	require.NoError(t, Db().Model(&PhotoLabel{}).Where("photo_id = ? AND label_src = ?", photo.ID, SrcCaption).Count(&count).Error)
	assert.Zero(t, count)

	changed, err = photo.ResetCaption(SrcVision)
	require.NoError(t, err)
	assert.False(t, changed)
}

func TestPhotoResetLabels(t *testing.T) {
	photo := createTestPhoto(t)

	label := NewLabel(fmt.Sprintf("reset-label-%s", rnd.GenerateUID(LabelUID)), 10)
	require.NoError(t, label.Create())

	t.Cleanup(func() {
		_ = Db().Delete(&PhotoLabel{}, "photo_id = ?", photo.ID).Error
		_ = Db().Delete(label).Error
	})

	require.NotNil(t, FirstOrCreatePhotoLabel(NewPhotoLabel(photo.ID, label.ID, 10, SrcOllama)))

	removed, err := photo.ResetLabels(SrcOllama)
	require.NoError(t, err)
	assert.EqualValues(t, 1, removed)

	var count int
	require.NoError(t, Db().Model(&PhotoLabel{}).Where("photo_id = ? AND label_src = ?", photo.ID, SrcOllama).Count(&count).Error)
	assert.Zero(t, count)

	removed, err = photo.ResetLabels(SrcVision)
	require.NoError(t, err)
	assert.Zero(t, removed)
}

// createTestPhoto builds an isolated photo row so reset tests can mutate captions
// and labels without relying on shared fixtures.
func createTestPhoto(t *testing.T) Photo {
	photo := NewUserPhoto(false, "")
	now := time.Now()
	photo.TakenAt = now
	photo.TakenAtLocal = now
	photo.PhotoPath = "test-reset"
	photo.PhotoName = fmt.Sprintf("%s.jpg", rnd.GenerateUID(PhotoUID))
	photo.OriginalName = photo.PhotoName

	require.NoError(t, Db().Create(&photo).Error)

	t.Cleanup(func() {
		_ = Db().Delete(&PhotoLabel{}, "photo_id = ?", photo.ID).Error
		_ = Db().Delete(&Photo{}, photo.ID).Error
	})

	return photo
}
