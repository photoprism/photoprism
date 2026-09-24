package entity

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
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
	t.Run("MakeAsPartOfModel", func(t *testing.T) {
		// The make is only removed from the model as a whole word, also after normalizing it.
		for _, c := range [][3]string{
			{"Nikon", "Nikonos V", "NIKON Nikonos V"},
			{"Canon", "Canonet QL17 GIII", "Canon Canonet QL17 GIII"},
			{"Rollei", "Rolleiflex 2.8F", "Rollei Rolleiflex 2.8F"},
			{"Leica", "Leicaflex SL", "Leica Leicaflex SL"},
			{"Yashica", "Yashica-Mat 124G", "Yashica Yashica-Mat 124G"},
		} {
			camera := NewCamera(c[0], c[1])
			assert.Equal(t, c[1], camera.CameraModel)
			assert.Equal(t, c[2], camera.CameraName)
		}
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

func TestAddCamera(t *testing.T) {
	// removeCamera deletes a test camera so that each case starts from a clean state.
	removeCamera := func(slug string) {
		cameraCache.Delete(slug)
		assert.NoError(t, UnscopedDb().Delete(&Camera{}, "camera_slug = ?", slug).Error)
	}

	t.Run("Created", func(t *testing.T) {
		slug := NewCamera("Minolta", "X-700").CameraSlug
		removeCamera(slug)
		t.Cleanup(func() { removeCamera(slug) })

		result, created, err := AddCamera("  Minolta ", " X-700  ")

		assert.NoError(t, err)
		assert.True(t, created)

		if result == nil {
			t.Fatal("result must not be nil")
		}

		assert.NotZero(t, result.ID)
		assert.Equal(t, slug, result.CameraSlug)
		assert.Equal(t, "Minolta X-700", result.CameraName)
		assert.Equal(t, SrcManual, result.CameraSrc)

		// The source must be persisted, not only set on the returned struct.
		found := Camera{}
		assert.NoError(t, Db().First(&found, "id = ?", result.ID).Error)
		assert.Equal(t, SrcManual, found.CameraSrc)

		// Adding the same camera again reports the existing record instead of a duplicate.
		again, created, err := AddCamera("Minolta", "X-700")
		assert.NoError(t, err)
		assert.False(t, created)
		assert.Equal(t, result.ID, again.ID)
	})
	t.Run("ExistingBySlug", func(t *testing.T) {
		fixture := CameraFixtures.Get("canon-eos-7d")
		t.Cleanup(func() {
			FlushCameraCache()
			assert.NoError(t, UnscopedDb().Save(CameraFixtures.Pointer("canon-eos-7d")).Error)
		})

		result, created, err := AddCamera(fixture.CameraMake, fixture.CameraModel)

		assert.NoError(t, err)
		assert.False(t, created)
		assert.Equal(t, fixture.ID, result.ID)
		assert.Equal(t, SrcManual, result.CameraSrc)

		found := Camera{}
		assert.NoError(t, Db().First(&found, "id = ?", fixture.ID).Error)
		assert.Equal(t, SrcManual, found.CameraSrc)
	})
	t.Run("ExistingRenamed", func(t *testing.T) {
		fixture := CameraFixtures.Get("canon-eos-7d")
		t.Cleanup(func() {
			FlushCameraCache()
			assert.NoError(t, UnscopedDb().Save(CameraFixtures.Pointer("canon-eos-7d")).Error)
		})

		// A renamed record keeps its slug, so it can only be found by make and model.
		renamed := Camera{}
		assert.NoError(t, Db().First(&renamed, "id = ?", fixture.ID).Error)
		assert.NoError(t, renamed.UpdateMakeModel("Minolta", "XD-7"))
		assert.Equal(t, fixture.CameraSlug, renamed.CameraSlug)
		assert.NotEqual(t, NewCamera("Minolta", "XD-7").CameraSlug, renamed.CameraSlug)

		result, created, err := AddCamera("Minolta", "XD-7")

		assert.NoError(t, err)
		assert.False(t, created)
		assert.Equal(t, fixture.ID, result.ID)
	})
	t.Run("EmptyMake", func(t *testing.T) {
		result, created, err := AddCamera("  ", "X-700")
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.ErrorContains(t, err, "make and model must not be empty")
		assert.False(t, created)
		assert.Nil(t, result)
	})
	t.Run("EmptyModel", func(t *testing.T) {
		result, created, err := AddCamera("Minolta", "")
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.ErrorContains(t, err, "make and model must not be empty")
		assert.False(t, created)
		assert.Nil(t, result)
	})
	t.Run("ModelSameAsMake", func(t *testing.T) {
		result, created, err := AddCamera("Zenit", "Zenit")
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.ErrorContains(t, err, "model must not be empty after removing the make")
		assert.False(t, created)
		assert.Nil(t, result)
	})
	t.Run("UnknownSlug", func(t *testing.T) {
		unknown := Camera{}
		assert.NoError(t, Db().First(&unknown, "id = ?", UnknownCamera.ID).Error)

		// These inputs normalize to the slug of the shared placeholder, which must stay untouched.
		for _, input := range [][2]string{{"ZZ", "."}, {"!", "zz"}, {"-", "zz"}} {
			result, created, err := AddCamera(input[0], input[1])
			assert.ErrorIs(t, err, ErrInvalidValue, "%q", input)
			assert.False(t, created)
			assert.Nil(t, result)
		}

		found := Camera{}
		assert.NoError(t, Db().First(&found, "id = ?", UnknownCamera.ID).Error)
		assert.Equal(t, unknown.CameraSrc, found.CameraSrc)
	})
	t.Run("ExistingPurged", func(t *testing.T) {
		slug := NewCamera("Zenit", "TTL").CameraSlug
		removeCamera(slug)
		t.Cleanup(func() { removeCamera(slug) })

		// A discovered orphan that is purged between the lookup and marking it is created again.
		orphan := FirstOrCreateCamera(NewCamera("Zenit", "TTL"))
		assert.NotZero(t, orphan.ID)
		removeCamera(slug)
		assert.ErrorIs(t, orphan.markManual(), gorm.ErrRecordNotFound)

		result, created, err := AddCamera("Zenit", "TTL")
		assert.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, SrcManual, result.CameraSrc)
	})
}

func TestCamera_markManual(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fixture := CameraFixtures.Get("canon-eos-5d")
		t.Cleanup(func() {
			FlushCameraCache()
			assert.NoError(t, UnscopedDb().Save(CameraFixtures.Pointer("canon-eos-5d")).Error)
		})

		m := Camera{}
		assert.NoError(t, Db().First(&m, "id = ?", fixture.ID).Error)
		assert.Empty(t, m.CameraSrc)
		assert.NoError(t, m.markManual())
		assert.Equal(t, SrcManual, m.CameraSrc)

		found := Camera{}
		assert.NoError(t, Db().First(&found, "id = ?", fixture.ID).Error)
		assert.Equal(t, SrcManual, found.CameraSrc)

		// Marking it again is a no-op.
		assert.NoError(t, m.markManual())
	})
	t.Run("NoPrimaryKey", func(t *testing.T) {
		m := Camera{CameraSlug: "no-primary-key"}
		assert.Error(t, m.markManual())
		assert.Empty(t, m.CameraSrc)
	})
}

func TestCamera_UpdateMakeModelUnknown(t *testing.T) {
	m := UnknownCamera
	assert.NotZero(t, m.ID)
	assert.EqualError(t, m.UpdateMakeModel("Minolta", "X-700"), "unknown camera cannot be changed")
	assert.Equal(t, UnknownCamera.CameraName, m.CameraName)
}
