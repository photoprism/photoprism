package workers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestVisionReset(t *testing.T) {
	conf := config.NewMinimalTestConfigWithDb("workers-vision-reset", t.TempDir())

	worker := NewVision(conf)
	fixture := entity.PhotoFixtures.Get("VisionResetTarget")
	require.NotEmpty(t, fixture.PhotoUID)

	landscape := entity.LabelFixtures.Get("landscape")

	photo := entity.FindPhoto(fixture)
	require.NotNil(t, photo)

	targetID := photo.ID
	targetUID := photo.PhotoUID

	t.Cleanup(func() {
		entity.FlushPhotoLabelCache()
	})

	require.NotNil(t, entity.FirstOrCreatePhotoLabel(entity.NewPhotoLabel(targetID, landscape.ID, 20, entity.SrcOllama)))

	require.NoError(t, entity.Db().Model(&entity.Photo{}).Where("id = ?", targetID).Updates(entity.Values{
		"PhotoCaption": "Reset caption",
		"CaptionSrc":   entity.SrcOllama,
	}).Error)

	entity.FlushPhotoLabelCache()

	filter := "uid:" + targetUID

	err := worker.Reset(filter, 10, []string{vision.ModelTypeCaption, vision.ModelTypeLabels}, entity.SrcOllama)
	require.NoError(t, err)

	refreshed := entity.FindPhoto(fixture)
	require.NotNil(t, refreshed)

	assert.Equal(t, "", refreshed.PhotoCaption)
	assert.Equal(t, "", refreshed.CaptionSrc)

	var labelCount int
	require.NoError(t, entity.Db().Model(&entity.PhotoLabel{}).Where("photo_id = ? AND label_src = ?", targetID, entity.SrcOllama).Count(&labelCount).Error)
	assert.Zero(t, labelCount)
}
