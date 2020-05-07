package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCamera_FirstOrCreate(t *testing.T) {
	t.Run("iphone-se", func(t *testing.T) {
		camera := NewCamera("iPhone SE", "Apple")
		camera.FirstOrCreate()
		assert.GreaterOrEqual(t, camera.ID, uint(1))
	})
}

func TestNewCamera(t *testing.T) {
	t.Run("unknown camera", func(t *testing.T) {
		camera := NewCamera("", "Nikon")

		assert.Equal(t, &UnknownCamera, camera)
	})
	t.Run("model EOS 6D make Canon", func(t *testing.T) {
		camera := NewCamera("EOS 6D", "Canon")

		expected := &Camera{
			CameraModel: "EOS 6D",
			CameraMake:  "Canon",
			CameraSlug:  "canon-eos-6d",
		}
		assert.Equal(t, expected, camera)
	})
	t.Run("model with prefix make Panasonic", func(t *testing.T) {
		camera := NewCamera("Panasonic Lumix", "Panasonic")

		expected := &Camera{
			CameraModel: "Lumix",
			CameraMake:  "Panasonic",
			CameraSlug:  "panasonic-lumix",
		}
		assert.Equal(t, expected, camera)
	})
	t.Run("model TG-4 make Unknown", func(t *testing.T) {
		camera := NewCamera("TG-4", "")

		expected := &Camera{
			CameraModel: "TG-4",
			CameraMake:  "",
			CameraSlug:  "tg-4",
		}
		assert.Equal(t, expected, camera)
	})
	t.Run("model Unknown make Unknown", func(t *testing.T) {
		camera := NewCamera("", "")

		assert.Equal(t, &UnknownCamera, camera)
	})
}

func TestCamera_String(t *testing.T) {
	t.Run("model XXX make Nikon", func(t *testing.T) {
		camera := NewCamera("XXX", "Nikon")
		cameraString := camera.String()
		assert.Equal(t, "Nikon XXX", cameraString)
	})
	t.Run("model XXX make Unknown", func(t *testing.T) {
		camera := NewCamera("XXX", "")
		cameraString := camera.String()
		assert.Equal(t, "XXX", cameraString)
	})
	t.Run("model Unknown make XXX", func(t *testing.T) {
		camera := NewCamera("", "test")
		cameraString := camera.String()
		assert.Equal(t, "Unknown", cameraString)
	})
	t.Run("model Unknown make Unknown", func(t *testing.T) {
		camera := NewCamera("", "")
		cameraString := camera.String()
		assert.Equal(t, "Unknown", cameraString)
	})
}
