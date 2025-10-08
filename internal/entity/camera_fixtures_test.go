package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCameraMap_Get(t *testing.T) {
	t.Run("GetExistingCamera", func(t *testing.T) {
		r := CameraFixtures.Get("apple-iphone-se")
		assert.Equal(t, uint(0xf4240), r.ID)
		assert.Equal(t, "iPhone SE", r.CameraModel)
		assert.IsType(t, Camera{}, r)
	})
	t.Run("GetNotExistingCamera", func(t *testing.T) {
		r := CameraFixtures.Get("fantasy Cam")
		assert.Equal(t, "fantasy-cam", r.CameraSlug)
		assert.IsType(t, Camera{}, r)
	})
}

func TestCameraMap_Pointer(t *testing.T) {
	t.Run("GetExistingCameraPointer", func(t *testing.T) {
		r := CameraFixtures.Pointer("apple-iphone-se")
		assert.Equal(t, uint(0xf4240), r.ID)
		assert.Equal(t, "iPhone SE", r.CameraModel)
		assert.IsType(t, &Camera{}, r)
	})
	t.Run("GetNotExistingCameraPointer", func(t *testing.T) {
		r := CameraFixtures.Pointer("GOPRO")
		assert.Equal(t, "gopro", r.CameraSlug)
		assert.IsType(t, &Camera{}, r)
	})
}
