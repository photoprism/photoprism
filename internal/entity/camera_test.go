package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
)

func TestFirstOrCreateCamera(t *testing.T) {
	t.Run("UnknownCamera", func(t *testing.T) {
		m := UnknownCamera

		assert.Equal(t, uint(1), m.ID)
		assert.Equal(t, UnknownID, m.CameraSlug)

		result := FirstOrCreateCamera(&m)

		if result == nil {
			t.Fatal("result must not be nil")
		}

		assert.Equal(t, uint(1), m.ID)
		assert.Equal(t, UnknownID, m.CameraSlug)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, UnknownID, result.CameraSlug)
	})
	t.Run("ExistingCamera", func(t *testing.T) {
		camera := NewCamera("Apple", "iPhone SE")

		result := FirstOrCreateCamera(camera)

		if result == nil {
			t.Fatal("result must not be nil")
		}

		assert.GreaterOrEqual(t, result.ID, uint(1))
	})
	t.Run("NotExistingCamera", func(t *testing.T) {
		camera := &Camera{ID: 10000000, CameraSlug: "camera-slug"}

		result := FirstOrCreateCamera(camera)

		if result == nil {
			t.Fatal("result must not be nil")
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
	t.Run("IPhone", func(t *testing.T) {
		camera := NewCamera(MakeApple, ModelIPhone)
		assert.Equal(t, CameraTypePhone, camera.CameraType)
		assert.Equal(t, MakeApple+" "+ModelIPhone, camera.CameraName)
		assert.Equal(t, MakeApple, camera.CameraMake)
		assert.Equal(t, ModelIPhone, camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.True(t, camera.Mobile())
	})
	t.Run("IPad", func(t *testing.T) {
		camera := NewCamera(MakeApple, ModelIPad)
		assert.Equal(t, CameraTypeTablet, camera.CameraType)
		assert.Equal(t, MakeApple+" "+ModelIPad, camera.CameraName)
		assert.Equal(t, MakeApple, camera.CameraMake)
		assert.Equal(t, ModelIPad, camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.True(t, camera.Mobile())
	})
	t.Run("IPadAir", func(t *testing.T) {
		camera := NewCamera(MakeApple, ModelIPadAir)
		assert.Equal(t, CameraTypeTablet, camera.CameraType)
		assert.Equal(t, MakeApple, camera.CameraMake)
		assert.Equal(t, ModelIPadAir, camera.CameraModel)
		assert.False(t, camera.Scanner())
		assert.True(t, camera.Mobile())
	})
	t.Run("IPadPro", func(t *testing.T) {
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

func TestCamera_UpdateMakeModel(t *testing.T) {
	t.Run("ExistingCamera", func(t *testing.T) {
		fixture := "canon-eos-7d"
		camera := NewCamera(CameraFixtures.Get(fixture).CameraMake, CameraFixtures.Get(fixture).CameraModel)

		result := FirstOrCreateCamera(camera)

		defer assert.NoError(t, UnscopedDb().Save(CameraFixtures.Pointer(fixture)).Error)
		makeName := "Pentax"
		modelName := "K-1"
		err := result.UpdateMakeModel(makeName, modelName)
		assert.NoError(t, err)
		assert.Equal(t, CameraFixtures.Get(fixture).ID, result.ID)
		assert.Equal(t, CameraMakes[makeName], result.CameraMake)
		assert.Equal(t, modelName, result.CameraModel)
		assert.Equal(t, CameraFixtures.Get(fixture).CameraSlug, result.CameraSlug) // Slug is preserved across renames.
		assert.Equal(t, CameraMakes[makeName]+" "+modelName, result.CameraName)
	})
	t.Run("NewCamera", func(t *testing.T) {
		setup := NewCamera("", "9 99")
		camera := FirstOrCreateCamera(setup)
		defer assert.NoError(t, UnscopedDb().Delete(&Camera{}, "id = ?", camera.ID).Error)
		makeName := "Pentax"
		modelName := "K-1"
		err := camera.UpdateMakeModel(makeName, modelName)
		assert.NoError(t, err)
		assert.Equal(t, CameraMakes[makeName], camera.CameraMake)
		assert.Equal(t, modelName, camera.CameraModel)
		assert.Equal(t, "9-99", camera.CameraSlug) // Slug is preserved across renames.
		assert.Equal(t, CameraMakes[makeName]+" "+modelName, camera.CameraName)
	})
	t.Run("NotExistingCamera", func(t *testing.T) {
		camera := NewCamera("", "9 98")
		err := camera.UpdateMakeModel("Pentax", "K-3")
		assert.Error(t, err)
	})
	t.Run("EmptyMake", func(t *testing.T) {
		camera := &Camera{ID: CameraFixtures.Get("canon-eos-7d").ID, CameraMake: "Canon", CameraModel: "EOS 7D", CameraName: "Canon EOS 7D", CameraSlug: "canon-eos-7d"}
		err := camera.UpdateMakeModel("  ", "EOS 7D")
		assert.Error(t, err)
		// The guard returns before any mutation, so existing values must be untouched.
		assert.Equal(t, "Canon", camera.CameraMake)
		assert.Equal(t, "EOS 7D", camera.CameraModel)
	})
	t.Run("EmptyModel", func(t *testing.T) {
		camera := &Camera{ID: CameraFixtures.Get("canon-eos-7d").ID, CameraMake: "Canon", CameraModel: "EOS 7D", CameraName: "Canon EOS 7D", CameraSlug: "canon-eos-7d"}
		err := camera.UpdateMakeModel("Canon", "")
		assert.Error(t, err)
		assert.Equal(t, "Canon", camera.CameraMake)
		assert.Equal(t, "EOS 7D", camera.CameraModel)
	})
}

// TestCamera_EntityEvents pins the camera content-channel payloads to the UID-only
// invariant: cameras.created/updated carry a []string of stable slugs, never entity
// fields, and an update does not republish the camera count.
func TestCamera_EntityEvents(t *testing.T) {
	t.Run("CreatedPublishesSlugOnly", func(t *testing.T) {
		m := NewCamera("Acme", "Test Camera 6789")

		// Force the create branch to fire regardless of prior runs, -count>1, or cache state.
		removeTestCamera := func() {
			cameraCache.Delete(m.CameraSlug)
			assert.NoError(t, UnscopedDb().Delete(&Camera{}, "camera_slug = ?", m.CameraSlug).Error)
		}
		removeTestCamera()
		t.Cleanup(removeTestCamera)

		sub := event.Subscribe("cameras.created")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		camera := FirstOrCreateCamera(m)

		if camera == nil {
			t.Fatal("result must not be nil")
		}

		select {
		case msg := <-sub.Receiver:
			assert.Equal(t, "cameras.created", msg.Name)
			slugs, ok := msg.Fields["entities"].([]string)
			assert.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
			assert.Equal(t, []string{camera.CameraSlug}, slugs)
		case <-time.After(2 * time.Second):
			t.Fatal("expected one cameras.created event")
		}
	})
	t.Run("UpdatedPublishesSlugOnlyWithoutCount", func(t *testing.T) {
		fixture := "canon-eos-7d"
		camera := FirstOrCreateCamera(NewCamera(CameraFixtures.Get(fixture).CameraMake, CameraFixtures.Get(fixture).CameraModel))
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Save(CameraFixtures.Pointer(fixture)).Error) })

		updated := event.Subscribe("cameras.updated")
		t.Cleanup(func() { event.Unsubscribe(updated) })
		count := event.Subscribe("count.cameras")
		t.Cleanup(func() { event.Unsubscribe(count) })

		assert.NoError(t, camera.UpdateMakeModel("Pentax", "K-1"))
		// The slug must be preserved across a Make/Model rename so the published identity is stable.
		assert.Equal(t, CameraFixtures.Get(fixture).CameraSlug, camera.CameraSlug)

		select {
		case msg := <-updated.Receiver:
			assert.Equal(t, "cameras.updated", msg.Name)
			slugs, ok := msg.Fields["entities"].([]string)
			assert.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
			assert.Equal(t, []string{camera.CameraSlug}, slugs)
		case <-time.After(2 * time.Second):
			t.Fatal("expected one cameras.updated event")
		}
		// An update does not change the camera count, so no count event is published.
		select {
		case msg := <-count.Receiver:
			t.Fatalf("unexpected %s event on camera update", msg.Name)
		case <-time.After(200 * time.Millisecond):
		}
	})
}

func TestCamera_SaveForm(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fixture := "canon-eos-7d"
		camera := FirstOrCreateCamera(NewCamera(CameraFixtures.Get(fixture).CameraMake, CameraFixtures.Get(fixture).CameraModel))
		defer assert.NoError(t, UnscopedDb().Save(CameraFixtures.Pointer(fixture)).Error)
		err := camera.SaveForm(&form.Camera{CameraMake: "Pentax", CameraModel: "K-1"})
		assert.NoError(t, err)
		assert.Equal(t, CameraMakes["Pentax"], camera.CameraMake)
		assert.Equal(t, "K-1", camera.CameraModel)
	})
	t.Run("NilForm", func(t *testing.T) {
		camera := &Camera{ID: CameraFixtures.Get("canon-eos-7d").ID}
		assert.Error(t, camera.SaveForm(nil))
	})
	t.Run("EmptyMake", func(t *testing.T) {
		camera := &Camera{ID: CameraFixtures.Get("canon-eos-7d").ID}
		assert.Error(t, camera.SaveForm(&form.Camera{CameraMake: "", CameraModel: "K-1"}))
	})
}
