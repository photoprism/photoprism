package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCamera(t *testing.T) {
	t.Run("model Unknown make Nikon", func(t *testing.T) {
		camera := NewCamera("", "Nikon")

		expected := &Camera{
			CameraModel: "Unknown",
			CameraMake:  "Nikon",
			CameraSlug:  "nikon-unknown",
		}
		assert.Equal(t, expected, camera)
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

		expected := &Camera{
			CameraModel: "Unknown",
			CameraMake:  "",
			CameraSlug:  "unknown",
		}
		assert.Equal(t, expected, camera)
	})
}

/* TODO find way to initialize db independently from config
func TestCamera_FirstOrCreate(t *testing.T) {
	t.Run("model random make Nikon", func(t *testing.T) {
		currentTime := time.Now()
		modelName := currentTime.String()
		camera := NewCamera(modelName, "Nikon")
		assert.Equal(t, uint(0x0), camera.ID)
		c := config.NewTestConfig()
		camera.FirstOrCreate(c.Db())
		t.Log(camera.ID)
	})
}*/

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
	t.Run("model Unkown make Unknown", func(t *testing.T) {
		camera := NewCamera("", "")
		cameraString := camera.String()
		assert.Equal(t, "Unknown", cameraString)
	})
}
