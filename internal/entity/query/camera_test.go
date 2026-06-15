package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestFindCameraBySlug(t *testing.T) {
	t.Run("ExistingCamera", func(t *testing.T) {
		camera := FindCameraBySlug(entity.CameraFixtures.Get("canon-eos-7d").CameraSlug)
		assert.NotNil(t, camera)
		if camera != nil {
			assert.Equal(t, entity.CameraFixtures.Get("canon-eos-7d").ID, camera.ID)
			assert.Equal(t, entity.CameraFixtures.Get("canon-eos-7d").CameraModel, camera.CameraModel)
		}
	})
	t.Run("NotExistingCamera", func(t *testing.T) {
		camera := FindCameraBySlug("IAmNotValid")
		assert.Nil(t, camera)
	})
	t.Run("EmptySlug", func(t *testing.T) {
		camera := FindCameraBySlug("")
		assert.Nil(t, camera)
	})
}

func TestFindCameraByID(t *testing.T) {
	t.Run("ExistingCamera", func(t *testing.T) {
		camera := FindCameraByID(entity.CameraFixtures.Get("canon-eos-7d").ID)
		assert.NotNil(t, camera)
		if camera != nil {
			assert.Equal(t, entity.CameraFixtures.Get("canon-eos-7d").ID, camera.ID)
			assert.Equal(t, entity.CameraFixtures.Get("canon-eos-7d").CameraModel, camera.CameraModel)
		}
	})
	t.Run("NotExistingCamera", func(t *testing.T) {
		camera := FindCameraByID(99885541348)
		assert.Nil(t, camera)
	})
	t.Run("ZeroID", func(t *testing.T) {
		camera := FindCameraByID(0)
		assert.Nil(t, camera)
	})
}
