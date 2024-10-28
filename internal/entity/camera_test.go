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
			CameraMake:  MakeCanon,
			CameraModel: "EOS 6D",
			CameraType:  CameraTypeBody,
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
			CameraMake:  MakeNone,
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
	t.Run("Empty", func(t *testing.T) {
		camera := Camera{}
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
	t.Run("KODAKSlideNScan", func(t *testing.T) {
		camera := NewCamera("GCMC", "RODFS50")
		assert.Equal(t, MakeKodak+" "+ModelSlideNScan, camera.CameraName)
		assert.Equal(t, CameraTypeFilm, camera.CameraType)
		assert.Equal(t, MakeKodak, camera.CameraMake)
		assert.Equal(t, ModelSlideNScan, camera.CameraModel)
		assert.True(t, camera.Scanner())
		assert.False(t, camera.Mobile())
	})
}

func TestCamera_Mobile(t *testing.T) {
	t.Run("CanonEOSD30", func(t *testing.T) {
		camera := NewCamera(MakeCanon, "EOS D30")
		assert.Equal(t, CameraTypeBody, camera.CameraType)
		assert.Equal(t, MakeCanon+" EOS D30", camera.CameraName)
		assert.Equal(t, MakeCanon, camera.CameraMake)
		assert.Equal(t, "EOS D30", camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.False(t, camera.Mobile())
	})
	t.Run("CanonEOS6D", func(t *testing.T) {
		camera := NewCamera(MakeCanon, "EOS 6D")
		assert.Equal(t, CameraTypeBody, camera.CameraType)
		assert.Equal(t, MakeCanon+" EOS 6D", camera.CameraName)
		assert.Equal(t, MakeCanon, camera.CameraMake)
		assert.Equal(t, "EOS 6D", camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.False(t, camera.Mobile())
	})
	t.Run("CanonEOSR6", func(t *testing.T) {
		camera := NewCamera(MakeCanon, "EOS R6")
		assert.Equal(t, CameraTypeBody, camera.CameraType)
		assert.Equal(t, MakeCanon+" EOS R6", camera.CameraName)
		assert.Equal(t, MakeCanon, camera.CameraMake)
		assert.Equal(t, "EOS R6", camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.False(t, camera.Mobile())
	})
	t.Run("CanonCinema", func(t *testing.T) {
		camera := NewCamera(MakeCanon, "EOS C100 Mark II")
		assert.Equal(t, CameraTypeVideo, camera.CameraType)
		assert.Equal(t, MakeCanon+" EOS C100 Mark II", camera.CameraName)
		assert.Equal(t, MakeCanon, camera.CameraMake)
		assert.Equal(t, "EOS C100 Mark II", camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.False(t, camera.Mobile())
	})
	t.Run("iPhone", func(t *testing.T) {
		camera := NewCamera(MakeApple, ModelIPhone)
		assert.Equal(t, CameraTypePhone, camera.CameraType)
		assert.Equal(t, MakeApple+" "+ModelIPhone, camera.CameraName)
		assert.Equal(t, MakeApple, camera.CameraMake)
		assert.Equal(t, ModelIPhone, camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.True(t, camera.Mobile())
	})
	t.Run("iPad", func(t *testing.T) {
		camera := NewCamera(MakeApple, ModelIPad)
		assert.Equal(t, CameraTypeTablet, camera.CameraType)
		assert.Equal(t, MakeApple+" "+ModelIPad, camera.CameraName)
		assert.Equal(t, MakeApple, camera.CameraMake)
		assert.Equal(t, ModelIPad, camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.True(t, camera.Mobile())
	})
	t.Run("iPadAir", func(t *testing.T) {
		camera := NewCamera(MakeApple, ModelIPadAir)
		assert.Equal(t, CameraTypeTablet, camera.CameraType)
		assert.Equal(t, MakeApple, camera.CameraMake)
		assert.Equal(t, ModelIPadAir, camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.True(t, camera.Mobile())
	})
	t.Run("iPadPro", func(t *testing.T) {
		camera := NewCamera(MakeApple, ModelIPadPro)
		assert.Equal(t, CameraTypeTablet, camera.CameraType)
		assert.Equal(t, MakeApple, camera.CameraMake)
		assert.Equal(t, ModelIPadPro, camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.True(t, camera.Mobile())
	})
	t.Run("SamsungGalaxyS21", func(t *testing.T) {
		camera := NewCamera(MakeSamsung, "Galaxy S21")
		assert.Equal(t, CameraTypePhone, camera.CameraType)
		assert.Equal(t, MakeSamsung, camera.CameraMake)
		assert.Equal(t, "Galaxy S21", camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.True(t, camera.Mobile())
	})
	t.Run("SamsungGalaxyTab", func(t *testing.T) {
		camera := NewCamera(MakeSamsung, "Galaxy Tab")
		assert.Equal(t, MakeSamsung+" Galaxy Tab", camera.CameraName)
		assert.Equal(t, CameraTypeTablet, camera.CameraType)
		assert.Equal(t, MakeSamsung, camera.CameraMake)
		assert.Equal(t, "Galaxy Tab", camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.True(t, camera.Mobile())
	})
}
