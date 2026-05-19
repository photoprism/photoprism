package search

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
)

// createHomophoneSearchFixtures adds two colliding labels with distinct names
// and assigns them to separate fixture photos for search regression tests.
func createHomophoneSearchFixtures(t *testing.T) (first, second *entity.Label, photoA, photoB entity.Photo) {
	t.Helper()

	entity.FlushLabelCache()
	entity.FlushPhotoLabelCache()

	photoA = entity.PhotoFixtures.Get("Photo09")
	photoB = entity.PhotoFixtures.Get("Photo10")

	first = entity.FirstOrCreateLabel(entity.NewLabel("问", 0))
	second = entity.FirstOrCreateLabel(entity.NewLabel("吻", 0))

	require.NotNil(t, first)
	require.NotNil(t, second)

	relationA := entity.FirstOrCreatePhotoLabel(entity.NewPhotoLabel(photoA.ID, first.ID, 10, entity.SrcManual))
	relationB := entity.FirstOrCreatePhotoLabel(entity.NewPhotoLabel(photoB.ID, second.ID, 10, entity.SrcManual))

	require.NotNil(t, relationA)
	require.NotNil(t, relationB)

	t.Cleanup(func() {
		_ = entity.Db().Where("photo_id = ? AND label_id = ?", photoA.ID, first.ID).Delete(&entity.PhotoLabel{}).Error
		_ = entity.Db().Where("photo_id = ? AND label_id = ?", photoB.ID, second.ID).Delete(&entity.PhotoLabel{}).Error
		_ = entity.Db().Unscoped().Delete(second).Error
		_ = entity.Db().Unscoped().Delete(first).Error
		entity.FlushLabelCache()
		entity.FlushPhotoLabelCache()
	})

	return first, second, photoA, photoB
}
