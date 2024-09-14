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
		camera := NewCamera("Apple", "iPhone SE")

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
	t.Run("Unknown", func(t *testing.T) {
		camera := NewCamera("", "")

		assert.Equal(t, &UnknownCamera, camera)
	})
	t.Run("CanonEOS6D", func(t *testing.T) {
		camera := NewCamera("Canon", "EOS 6D")

		expected := &Camera{
			CameraSlug:  "canon-eos-6d",
			CameraName:  "Canon EOS 6D",
			CameraMake:  "Canon",
			CameraModel: "EOS 6D",
		}

		assert.Equal(t, expected, camera)
	})
	t.Run("PanasonicLumix", func(t *testing.T) {
		camera := NewCamera("Panasonic", "Panasonic Lumix")

		expected := &Camera{
			CameraSlug:  "panasonic-lumix",
			CameraName:  "Panasonic Lumix",
			CameraMake:  "Panasonic",
			CameraModel: "Lumix",
		}

		assert.Equal(t, expected, camera)
	})
	t.Run("TG4", func(t *testing.T) {
		camera := NewCamera("", "TG-4")

		expected := &Camera{
			CameraSlug:  "tg-4",
			CameraName:  "TG-4",
			CameraMake:  "",
			CameraModel: "TG-4",
		}

		assert.Equal(t, expected, camera)
	})
	t.Run("Olympus", func(t *testing.T) {
		camera := NewCamera("OLYMPUS OPTICAL CO.,LTD", "")

		assert.Equal(t, "olympus", camera.CameraSlug)
		assert.Equal(t, "Olympus", camera.CameraName)
		assert.Equal(t, "Olympus", camera.CameraMake)
		assert.Equal(t, "", camera.CameraModel)
	})
	t.Run("HuaweiP30", func(t *testing.T) {
		camera := NewCamera("Huawei", "ELE-AL00")

		assert.Equal(t, "huawei-p30", camera.CameraSlug)
		assert.Equal(t, "HUAWEI P30", camera.CameraName)
		assert.Equal(t, "HUAWEI", camera.CameraMake)
		assert.Equal(t, "P30", camera.CameraModel)
	})
}

func TestCamera_String(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		camera := NewCamera("", "")
		cameraString := camera.String()
		assert.Equal(t, "Unknown", cameraString)
	})
	t.Run("Nikon", func(t *testing.T) {
		camera := NewCamera("Nikon", "foo")
		cameraString := camera.String()
		assert.Equal(t, "'NIKON foo'", cameraString)
	})
	t.Run("Foo", func(t *testing.T) {
		camera := NewCamera("", "Foo")
		cameraString := camera.String()
		assert.Equal(t, "Foo", cameraString)
	})
	t.Run("Test", func(t *testing.T) {
		camera := NewCamera("test", "")
		cameraString := camera.String()
		assert.Equal(t, "test", cameraString)
	})
}

func TestCamera_Scanner(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		camera := NewCamera("", "")
		assert.False(t, camera.Scanner())
	})
	t.Run("Foo", func(t *testing.T) {
		camera := NewCamera("foo", "")
		assert.False(t, camera.Scanner())
	})
	t.Run("NikonFoo", func(t *testing.T) {
		camera := NewCamera("Nikon", "Foo")
		assert.False(t, camera.Scanner())
	})
	t.Run("MSScanner", func(t *testing.T) {
		camera := NewCamera("", "MS Scanner")
		assert.True(t, camera.Scanner())
	})
}
