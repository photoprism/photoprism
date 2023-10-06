package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirstOrCreateCamera(t *testing.T) {
	t.Run("UnknownCamera", func(t *testing.T) {
		m := UnknownCamera

		assert.Equal(t, uint(1), m.ID)
		assert.Equal(t, UnknownID, m.CameraSlug)

		result := FirstOrCreateCamera(&m)

		if result == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, uint(1), m.ID)
		assert.Equal(t, UnknownID, m.CameraSlug)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, UnknownID, result.CameraSlug)
	})
	t.Run("existing camera", func(t *testing.T) {
		camera := NewCamera("iPhone SE", "Apple")

		result := FirstOrCreateCamera(camera)

		if result == nil {
			t.Fatal("result should not be nil")
		}

		assert.GreaterOrEqual(t, result.ID, uint(1))
	})
	t.Run("not existing camera", func(t *testing.T) {
		camera := &Camera{ID: 10000000, CameraSlug: "camera-slug"}

		result := FirstOrCreateCamera(camera)

		if result == nil {
			t.Fatal("result should not be nil")
		}

		assert.GreaterOrEqual(t, result.ID, uint(1))
	})
}

func TestNewCamera(t *testing.T) {
	t.Run("unknown camera", func(t *testing.T) {
		camera := NewCamera("", "")

		assert.Equal(t, &UnknownCamera, camera)
	})
	t.Run("model EOS 6D make Canon", func(t *testing.T) {
		camera := NewCamera("EOS 6D", "Canon")

		expected := &Camera{
			CameraSlug:  "canon-eos-6d",
			CameraName:  "Canon EOS 6D",
			CameraMake:  "Canon",
			CameraModel: "EOS 6D",
		}

		assert.Equal(t, expected, camera)
	})
	t.Run("model with prefix make Panasonic", func(t *testing.T) {
		camera := NewCamera("Panasonic Lumix", "Panasonic")

		expected := &Camera{
			CameraSlug:  "panasonic-lumix",
			CameraName:  "Panasonic Lumix",
			CameraMake:  "Panasonic",
			CameraModel: "Lumix",
		}

		assert.Equal(t, expected, camera)
	})
	t.Run("model TG-4 make Unknown", func(t *testing.T) {
		camera := NewCamera("TG-4", "")

		expected := &Camera{
			CameraSlug:  "tg-4",
			CameraName:  "TG-4",
			CameraMake:  "",
			CameraModel: "TG-4",
		}

		assert.Equal(t, expected, camera)
	})
	t.Run("model Unknown make Unknown", func(t *testing.T) {
		camera := NewCamera("", "")

		assert.Equal(t, &UnknownCamera, camera)
	})

	t.Run("OLYMPUS", func(t *testing.T) {
		camera := NewCamera("", "OLYMPUS OPTICAL CO.,LTD")

		assert.Equal(t, "olympus", camera.CameraSlug)
		assert.Equal(t, "Olympus", camera.CameraName)
		assert.Equal(t, "Olympus", camera.CameraMake)
		assert.Equal(t, "", camera.CameraModel)
	})

	t.Run("P30", func(t *testing.T) {
		camera := NewCamera("ELE-AL00", "Huawei")

		assert.Equal(t, "huawei-p30", camera.CameraSlug)
		assert.Equal(t, "HUAWEI P30", camera.CameraName)
		assert.Equal(t, "HUAWEI", camera.CameraMake)
		assert.Equal(t, "P30", camera.CameraModel)
	})
}

func TestCamera_String(t *testing.T) {
	t.Run("model XXX make Nikon", func(t *testing.T) {
		camera := NewCamera("XXX", "Nikon")
		cameraString := camera.String()
		assert.Equal(t, "'NIKON XXX'", cameraString)
	})
	t.Run("model XXX make Unknown", func(t *testing.T) {
		camera := NewCamera("XXX", "")
		cameraString := camera.String()
		assert.Equal(t, "XXX", cameraString)
	})
	t.Run("model Unknown make XXX", func(t *testing.T) {
		camera := NewCamera("", "test")
		cameraString := camera.String()
		assert.Equal(t, "test", cameraString)
	})
	t.Run("model Unknown make Unknown", func(t *testing.T) {
		camera := NewCamera("", "")
		cameraString := camera.String()
		assert.Equal(t, "Unknown", cameraString)
	})
}

func TestCamera_Scanner(t *testing.T) {
	t.Run("model XXX make Nikon", func(t *testing.T) {
		camera := NewCamera("XXX", "Nikon")
		assert.False(t, camera.Scanner())
	})
	t.Run("MS Scanner", func(t *testing.T) {
		camera := NewCamera("MS Scanner", "")
		assert.True(t, camera.Scanner())
	})
	t.Run("model Unknown make XXX", func(t *testing.T) {
		camera := NewCamera("", "test")
		assert.False(t, camera.Scanner())
	})
	t.Run("model Unknown make Unknown", func(t *testing.T) {
		camera := NewCamera("", "")
		assert.False(t, camera.Scanner())
	})
}
