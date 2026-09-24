package entity

import (
	"strings"
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

// useCamera temporarily assigns the first photo to the camera and returns the photo ID.
func useCamera(t *testing.T, m *Camera) uint {
	t.Helper()

	photo := Photo{}
	assert.NoError(t, UnscopedDb().Order("id").First(&photo).Error)
	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Model(&Photo{}).Where("id = ?", photo.ID).UpdateColumn("camera_id", photo.CameraID).Error)
	})
	assert.NoError(t, UnscopedDb().Model(&Photo{}).Where("id = ?", photo.ID).UpdateColumn("camera_id", m.ID).Error)

	return photo.ID
}

// addCamera adds a camera for testing and removes it again when the test is done.
func addCamera(t *testing.T, makeName, modelName string) *Camera {
	t.Helper()

	slug := NewCamera(makeName, modelName).CameraSlug
	remove := func() {
		cameraCache.Delete(slug)
		assert.NoError(t, UnscopedDb().Delete(&Camera{}, "camera_slug = ?", slug).Error)
	}
	remove()
	t.Cleanup(remove)

	m, created, err := AddCamera(makeName, modelName)
	assert.NoError(t, err)
	assert.True(t, created)

	if m == nil {
		t.Fatal("camera must not be nil")
	}

	return m
}

func TestCamera_PhotoCount(t *testing.T) {
	t.Run("Unused", func(t *testing.T) {
		m := addCamera(t, "Zenit", "E")
		count, err := m.PhotoCount()
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
	t.Run("Used", func(t *testing.T) {
		m := addCamera(t, "Zenit", "12XP")
		useCamera(t, m)
		count, err := m.PhotoCount()
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
	t.Run("EmptyID", func(t *testing.T) {
		_, err := (&Camera{}).PhotoCount()
		assert.Error(t, err)
	})
}

func TestCamera_Delete(t *testing.T) {
	exists := func(t *testing.T, id uint) bool {
		var count int
		assert.NoError(t, UnscopedDb().Model(&Camera{}).Where("id = ?", id).Count(&count).Error)
		return count > 0
	}

	t.Run("Unused", func(t *testing.T) {
		m := addCamera(t, "Zenit", "E")

		deleted := event.Subscribe("cameras.deleted")
		t.Cleanup(func() { event.Unsubscribe(deleted) })

		reassigned, err := m.Delete(false)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), reassigned)
		assert.False(t, exists(t, m.ID))

		_, cached := cameraCache.Get(m.CameraSlug)
		assert.False(t, cached)

		select {
		case msg := <-deleted.Receiver:
			assert.Equal(t, []string{m.CameraSlug}, msg.Fields["entities"])
		case <-time.After(2 * time.Second):
			t.Fatal("expected one cameras.deleted event")
		}
	})
	t.Run("InUse", func(t *testing.T) {
		m := addCamera(t, "Zenit", "12XP")
		photoID := useCamera(t, m)

		reassigned, err := m.Delete(false)
		assert.ErrorIs(t, err, ErrInUse)
		assert.Equal(t, int64(0), reassigned)
		assert.True(t, exists(t, m.ID))

		photo := Photo{}
		assert.NoError(t, UnscopedDb().First(&photo, "id = ?", photoID).Error)
		assert.Equal(t, m.ID, photo.CameraID)
	})
	t.Run("Reassign", func(t *testing.T) {
		m := addCamera(t, "Zenit", "122")
		photoID := useCamera(t, m)

		reassigned, err := m.Delete(true)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), reassigned)
		assert.False(t, exists(t, m.ID))

		photo := Photo{}
		assert.NoError(t, UnscopedDb().First(&photo, "id = ?", photoID).Error)
		assert.Equal(t, UnknownCamera.ID, photo.CameraID)
	})
	t.Run("Unknown", func(t *testing.T) {
		m := UnknownCamera
		_, err := m.Delete(true)
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.True(t, exists(t, UnknownCamera.ID))
	})
	t.Run("NotFound", func(t *testing.T) {
		m := Camera{ID: 999999999, CameraSlug: "not-found"}
		_, err := m.Delete(false)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("EmptyID", func(t *testing.T) {
		_, err := (&Camera{CameraSlug: "empty-id"}).Delete(false)
		assert.Error(t, err)
	})
}

func TestFindCamerasByMakeModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := addCamera(t, "Zenit", "TTL")
		found := FindCamerasByMakeModel("  Zenit ", "Zenit TTL")

		if assert.Len(t, found, 1) {
			assert.Equal(t, m.ID, found[0].ID)
		}
	})
	t.Run("Renamed", func(t *testing.T) {
		fixture := CameraFixtures.Get("canon-eos-7d")
		t.Cleanup(func() {
			FlushCameraCache()
			assert.NoError(t, UnscopedDb().Save(CameraFixtures.Pointer("canon-eos-7d")).Error)
		})

		renamed := Camera{}
		assert.NoError(t, Db().First(&renamed, "id = ?", fixture.ID).Error)
		assert.NoError(t, renamed.UpdateMakeModel("Minolta", "XD-7"))

		// The old name still matches the slug, but must no longer find the renamed record.
		assert.Empty(t, FindCamerasByMakeModel(fixture.CameraMake, fixture.CameraModel))

		if found := FindCamerasByMakeModel("Minolta", "XD-7"); assert.Len(t, found, 1) {
			assert.Equal(t, fixture.ID, found[0].ID)
		}
	})
	t.Run("AsStored", func(t *testing.T) {
		// A record saved without normalizing its make is selected as stored, not its normalized twin.
		normalized := addCamera(t, "Pentax", "11")
		assert.Equal(t, "PENTAX", normalized.CameraMake)

		stored := Camera{CameraSlug: "as-stored-camera-test", CameraName: "Pentax 11", CameraMake: "Pentax", CameraModel: "11"}
		assert.NoError(t, UnscopedDb().Create(&stored).Error)
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&Camera{}, "id = ?", stored.ID).Error) })

		if found := FindCamerasByMakeModel("Pentax", "11"); assert.Len(t, found, 1) {
			assert.Equal(t, stored.ID, found[0].ID)
		}

		if found := FindCamerasByMakeModel("PENTAX", "11"); assert.Len(t, found, 1) {
			assert.Equal(t, normalized.ID, found[0].ID)
		}
	})
	t.Run("Duplicates", func(t *testing.T) {
		first := addCamera(t, "Zenit", "19")

		second := Camera{CameraSlug: "duplicate-camera-test", CameraName: first.CameraName, CameraMake: first.CameraMake, CameraModel: first.CameraModel}
		assert.NoError(t, UnscopedDb().Create(&second).Error)
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&Camera{}, "id = ?", second.ID).Error) })

		if found := FindCamerasByMakeModel("Zenit", "19"); assert.Len(t, found, 2) {
			assert.Equal(t, first.ID, found[0].ID)
			assert.Equal(t, second.ID, found[1].ID)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		assert.Empty(t, FindCamerasByMakeModel("Zenit", "Does Not Exist"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Empty(t, FindCamerasByMakeModel("", ""))
		assert.Empty(t, FindCamerasByMakeModel("ZZ", "."))
		assert.Empty(t, FindCamerasByMakeModel("", "Unknown"))
	})
	t.Run("ModelSameAsMake", func(t *testing.T) {
		assert.Empty(t, FindCamerasByMakeModel("Zenit", "Zenit"))
	})
}

func TestFindCamerasByMakeModelExact(t *testing.T) {
	t.Run("Exact", func(t *testing.T) {
		m := addCamera(t, "Zenit", "TTL")

		if found := findCamerasByMakeModel(m.CameraMake, m.CameraModel); assert.Len(t, found, 1) {
			assert.Equal(t, m.ID, found[0].ID)
		}
	})
	t.Run("CaseDiffers", func(t *testing.T) {
		m := addCamera(t, "Zenit", "TTL")

		// MariaDB collations match this in SQL, so the result must be filtered in Go.
		assert.Empty(t, findCamerasByMakeModel(strings.ToUpper(m.CameraMake), m.CameraModel))
	})
	t.Run("Placeholder", func(t *testing.T) {
		assert.Empty(t, findCamerasByMakeModel(UnknownCamera.CameraMake, UnknownCamera.CameraModel))
	})
}

func TestUnknownCameraID(t *testing.T) {
	placeholder := Camera{}
	assert.NoError(t, UnscopedDb().First(&placeholder, "camera_slug = ?", UnknownID).Error)

	t.Run("Initialized", func(t *testing.T) {
		id, err := unknownCameraID()
		assert.NoError(t, err)
		assert.Equal(t, placeholder.ID, id)
	})
	t.Run("Uninitialized", func(t *testing.T) {
		// The CLI does not initialize the placeholder on startup, so its ID must be read from the database.
		prev := UnknownCamera
		t.Cleanup(func() { UnknownCamera = prev })
		UnknownCamera.ID = 0
		FlushCameraCache()

		id, err := unknownCameraID()
		assert.NoError(t, err)
		assert.Equal(t, placeholder.ID, id)
	})
}

func TestFindExistingCamera(t *testing.T) {
	t.Run("BySlug", func(t *testing.T) {
		fixture := CameraFixtures.Get("canon-eos-7d")
		found := findExistingCamera(NewCamera(fixture.CameraMake, fixture.CameraModel))

		if found == nil {
			t.Fatal("camera must be found")
		}

		assert.Equal(t, fixture.ID, found.ID)
	})
	t.Run("ByMakeModel", func(t *testing.T) {
		fixture := CameraFixtures.Get("canon-eos-7d")
		t.Cleanup(func() {
			FlushCameraCache()
			assert.NoError(t, UnscopedDb().Save(CameraFixtures.Pointer("canon-eos-7d")).Error)
		})

		renamed := Camera{}
		assert.NoError(t, Db().First(&renamed, "id = ?", fixture.ID).Error)
		assert.NoError(t, renamed.UpdateMakeModel("Minolta", "XD-7"))

		found := findExistingCamera(NewCamera("Minolta", "XD-7"))

		if found == nil {
			t.Fatal("camera must be found")
		}

		assert.Equal(t, fixture.ID, found.ID)
	})
	t.Run("None", func(t *testing.T) {
		assert.Nil(t, findExistingCamera(NewCamera("Zenit", "Does Not Exist")))
	})
}

func TestCamera_DeleteEdgeCases(t *testing.T) {
	exists := func(t *testing.T, id uint) bool {
		var count int
		assert.NoError(t, UnscopedDb().Model(&Camera{}).Where("id = ?", id).Count(&count).Error)
		return count > 0
	}

	placeholder := Camera{}
	assert.NoError(t, UnscopedDb().First(&placeholder, "camera_slug = ?", UnknownID).Error)

	t.Run("OnlyTheRecord", func(t *testing.T) {
		m := addCamera(t, "Zenit", "TTL")
		other := addCamera(t, "Zenit", "11")

		_, err := m.Delete(false)
		assert.NoError(t, err)
		assert.False(t, exists(t, m.ID))
		assert.True(t, exists(t, other.ID))
	})
	t.Run("SoftDeletedPhoto", func(t *testing.T) {
		m := addCamera(t, "Zenit", "19")
		photoID := useCamera(t, m)

		// Archived and deleted pictures still reference the camera and must be counted and reassigned.
		deletedAt := time.Now().UTC()
		assert.NoError(t, UnscopedDb().Model(&Photo{}).Where("id = ?", photoID).UpdateColumn("deleted_at", &deletedAt).Error)
		t.Cleanup(func() {
			assert.NoError(t, UnscopedDb().Model(&Photo{}).Where("id = ?", photoID).UpdateColumn("deleted_at", nil).Error)
		})

		count, err := m.PhotoCount()
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		_, err = m.Delete(false)
		assert.ErrorIs(t, err, ErrInUse)

		reassigned, err := m.Delete(true)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), reassigned)

		photo := Photo{}
		assert.NoError(t, UnscopedDb().First(&photo, "id = ?", photoID).Error)
		assert.Equal(t, placeholder.ID, photo.CameraID)
	})
	t.Run("UninitializedPlaceholder", func(t *testing.T) {
		m := addCamera(t, "Zenit", "Photosniper")
		photoID := useCamera(t, m)

		// The CLI does not initialize the placeholder, so reassigning must not write ID 0.
		prev := UnknownCamera
		t.Cleanup(func() { UnknownCamera = prev })
		UnknownCamera.ID = 0
		FlushCameraCache()

		reassigned, err := m.Delete(true)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), reassigned)

		photo := Photo{}
		assert.NoError(t, UnscopedDb().First(&photo, "id = ?", photoID).Error)
		assert.Equal(t, placeholder.ID, photo.CameraID)
	})
	t.Run("PlaceholderID", func(t *testing.T) {
		m := Camera{ID: placeholder.ID, CameraSlug: "not-the-placeholder-slug"}
		_, err := m.Delete(true)
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.True(t, exists(t, placeholder.ID))
	})
	t.Run("PlaceholderSlug", func(t *testing.T) {
		m := addCamera(t, "Zenit", "EM")
		m.CameraSlug = UnknownID
		_, err := m.Delete(true)
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.True(t, exists(t, m.ID))
	})
}
