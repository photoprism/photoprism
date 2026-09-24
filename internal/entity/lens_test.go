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

func TestNewLens(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		lens := NewLens("", "")
		assert.Equal(t, UnknownID, lens.LensSlug)
		assert.Equal(t, "Unknown", lens.LensName)
		assert.Equal(t, "Unknown", lens.LensModel)
		assert.Equal(t, "", lens.LensMake)
		assert.Equal(t, UnknownLens.LensSlug, lens.LensSlug)
		assert.Equal(t, &UnknownLens, lens)
	})
	t.Run("Canon", func(t *testing.T) {
		lens := NewLens("Canon", "F500-99")
		assert.Equal(t, "canon-f500-99", lens.LensSlug)
		assert.Equal(t, "Canon F500-99", lens.LensName)
		assert.Equal(t, "F500-99", lens.LensModel)
		assert.Equal(t, "Canon", lens.LensMake)
	})
	t.Run("MakeAsPartOfModel", func(t *testing.T) {
		// The make is only removed from the model as a whole word.
		lens := NewLens("Helios", "Helios-44-2 58mm f/2")
		assert.Equal(t, "helios-helios-44-2-58mm-f-2", lens.LensSlug)
		assert.Equal(t, "Helios Helios-44-2 58mm f/2", lens.LensName)
		assert.Equal(t, "Helios-44-2 58mm f/2", lens.LensModel)
		assert.Equal(t, "Helios", lens.LensMake)
	})
	t.Run("MakeAsFirstWordOfModel", func(t *testing.T) {
		lens := NewLens("Helios", "Helios 44-2 58mm f/2")
		assert.Equal(t, "helios-44-2-58mm-f-2", lens.LensSlug)
		assert.Equal(t, "Helios 44-2 58mm f/2", lens.LensName)
		assert.Equal(t, "44-2 58mm f/2", lens.LensModel)
	})
	t.Run("IPhoneXs", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone XS back camera 4.25mm f/1.8")
		assert.Equal(t, "apple-iphone-xs-4-25mm-f-1-8", lens.LensSlug)
		assert.Equal(t, "Apple iPhone XS 4.25mm f/1.8", lens.LensName)
		assert.Equal(t, "iPhone XS 4.25mm f/1.8", lens.LensModel)
		assert.Equal(t, "Apple", lens.LensMake)
	})
	t.Run("IPhoneTwelveMini", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone 12 mini back dual wide camera 4.2mm f/1.6")
		assert.Equal(t, "apple-iphone-12-mini-4-2mm-f-1-6", lens.LensSlug)
		assert.Equal(t, "Apple iPhone 12 mini 4.2mm f/1.6", lens.LensName)
		assert.Equal(t, "iPhone 12 mini 4.2mm f/1.6", lens.LensModel)
		assert.Equal(t, "Apple", lens.LensMake)
	})
	t.Run("IPhoneTwelveUltraWide", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone 12 back dual wide camera 1.55mm f/2.4")
		assert.Equal(t, "apple-iphone-12-1-55mm-f-2-4", lens.LensSlug)
		assert.Equal(t, "Apple iPhone 12 1.55mm f/2.4", lens.LensName)
		assert.Equal(t, "iPhone 12 1.55mm f/2.4", lens.LensModel)
		assert.Equal(t, "Apple", lens.LensMake)
	})
	t.Run("IPhoneFourteenProMax", func(t *testing.T) {
		lens := NewLens("Apple", "iPhone 14 Pro Max back triple camera 9mm f/2.8")
		assert.Equal(t, "apple-iphone-14-pro-max-9mm-f-2-8", lens.LensSlug)
		assert.Equal(t, "Apple iPhone 14 Pro Max 9mm f/2.8", lens.LensName)
		assert.Equal(t, "iPhone 14 Pro Max 9mm f/2.8", lens.LensModel)
		assert.Equal(t, "Apple", lens.LensMake)
		assert.Equal(t, "apple-iphone-14-pro-max-9mm-f-2-8", lens.LensSlug)
	})
}

func TestLens_TableName(t *testing.T) {
	lens := NewLens("Canon", "F500-99")
	tableName := lens.TableName()
	assert.Equal(t, "lenses", tableName)
}

func TestLens_String(t *testing.T) {
	lens := NewLens("samsung", "F500-99")
	assert.Equal(t, "'Samsung F500-99'", lens.String())
}

func TestFirstOrCreateLens(t *testing.T) {
	t.Run("ExistingLens", func(t *testing.T) {
		fixture := "4.15mm-f/2.2"
		lens := NewLens(LensFixtures.Get(fixture).LensMake, "iPhone SE back camera 4.15mm f/2.2") // Use value that comes back from exiftool

		result := FirstOrCreateLens(lens)

		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, LensFixtures.Get(fixture).ID, result.ID)
			assert.Equal(t, LensFixtures.Get(fixture).LensMake, result.LensMake)
			assert.Equal(t, LensFixtures.Get(fixture).LensModel, result.LensModel)
			assert.Equal(t, LensFixtures.Get(fixture).LensSlug, result.LensSlug)
			assert.Equal(t, LensFixtures.Get(fixture).LensName, result.LensName)
		}
	})
	t.Run("ExistingLensWithOnlyModel", func(t *testing.T) {
		// This tests the Pentax lens situation
		fixture := "4-37"
		lens := NewLens(LensFixtures.Get(fixture).LensMake, LensFixtures.Get(fixture).LensModel)

		result := FirstOrCreateLens(lens)

		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, LensFixtures.Get(fixture).ID, result.ID)
			assert.Equal(t, LensFixtures.Get(fixture).LensMake, result.LensMake)
			assert.Equal(t, LensFixtures.Get(fixture).LensModel, result.LensModel)
			assert.Equal(t, LensFixtures.Get(fixture).LensSlug, result.LensSlug)
			assert.Equal(t, LensFixtures.Get(fixture).LensName, result.LensName)
		}
	})
	t.Run("NotExistingLens", func(t *testing.T) {
		lens := &Lens{}

		result := FirstOrCreateLens(lens)

		if result == nil {
			t.Fatal("result must not be nil")
		}
		assert.GreaterOrEqual(t, result.ID, uint(1))
	})
}

func TestLensUpdateMakeModel(t *testing.T) {
	t.Run("ExistingLens", func(t *testing.T) {
		fixture := "4-37"
		lens := NewLens(LensFixtures.Get(fixture).LensMake, LensFixtures.Get(fixture).LensModel)

		result := FirstOrCreateLens(lens)

		defer assert.NoError(t, UnscopedDb().Save(LensFixtures.Pointer(fixture)).Error)
		make := "Tamron"
		model := "Tamron SP AF 24-135mm F3.5-5.6 AD AL (190D)"
		err := result.UpdateMakeModel(make, model)
		assert.NoError(t, err)
		assert.Equal(t, LensFixtures.Get(fixture).ID, result.ID)
		assert.Equal(t, make, result.LensMake)
		assert.Equal(t, "SP AF 24-135mm F3.5-5.6 AD AL (190D)", result.LensModel) // NewLens strips Tamron from model to prevent double up in name
		assert.Equal(t, LensFixtures.Get(fixture).LensSlug, result.LensSlug)
		assert.Equal(t, model, result.LensName) // NewLens prevents Tamron Tamron ... as name
	})
	t.Run("NewLens", func(t *testing.T) {
		setup := NewLens("", "4 38")
		lens := FirstOrCreateLens(setup)
		defer assert.NoError(t, UnscopedDb().Delete(&Lens{}, "id = ?", lens.ID).Error)
		make := "Pentax"
		model := "smc PENTAX-FA 28-105mm F3.2-4.5 AL[IF]"
		err := lens.UpdateMakeModel(make, model)
		assert.NoError(t, err)
		assert.Equal(t, CameraMakes[make], lens.LensMake)
		assert.Equal(t, model, lens.LensModel)
		assert.Equal(t, "4-38", lens.LensSlug)
		assert.Equal(t, "PENTAX smc PENTAX-FA 28-105mm F3.2-4.5 AL[IF]", lens.LensName)
	})
	t.Run("NotExistingLens", func(t *testing.T) {
		lens := NewLens("", "4 39")
		err := lens.UpdateMakeModel("Pentax", "smc PENTAX-FA 31mm F1.8 AL Limited")
		assert.Error(t, err)
	})
	t.Run("EmptyMake", func(t *testing.T) {
		lens := &Lens{ID: LensFixtures.Get("lens-f-380").ID, LensMake: "Apple", LensModel: "F380", LensName: "Apple F380", LensSlug: "lens-f-380"}
		err := lens.UpdateMakeModel("  ", "F380")
		assert.Error(t, err)
		// The guard returns before any mutation, so existing values must be untouched.
		assert.Equal(t, "Apple", lens.LensMake)
		assert.Equal(t, "F380", lens.LensModel)
	})
	t.Run("EmptyModel", func(t *testing.T) {
		lens := &Lens{ID: LensFixtures.Get("lens-f-380").ID, LensMake: "Apple", LensModel: "F380", LensName: "Apple F380", LensSlug: "lens-f-380"}
		err := lens.UpdateMakeModel("Apple", "")
		assert.Error(t, err)
		assert.Equal(t, "Apple", lens.LensMake)
		assert.Equal(t, "F380", lens.LensModel)
	})
}

// TestLens_EntityEvents pins the lens content-channel payloads to the UID-only
// invariant: lenses.created/updated carry a []string of stable slugs, never entity
// fields, and an update does not republish the lens count.
func TestLens_EntityEvents(t *testing.T) {
	t.Run("CreatedPublishesSlugOnly", func(t *testing.T) {
		m := NewLens("Acme", "Test Lens 6789")

		// Force the create branch to fire regardless of prior runs, -count>1, or cache state.
		removeTestLens := func() {
			lensCache.Delete(m.LensSlug)
			assert.NoError(t, UnscopedDb().Delete(&Lens{}, "lens_slug = ?", m.LensSlug).Error)
		}
		removeTestLens()
		t.Cleanup(removeTestLens)

		sub := event.Subscribe("lenses.created")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		lens := FirstOrCreateLens(m)

		if lens == nil {
			t.Fatal("result must not be nil")
		}

		select {
		case msg := <-sub.Receiver:
			assert.Equal(t, "lenses.created", msg.Name)
			slugs, ok := msg.Fields["entities"].([]string)
			assert.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
			assert.Equal(t, []string{lens.LensSlug}, slugs)
		case <-time.After(2 * time.Second):
			t.Fatal("expected one lenses.created event")
		}
	})
	t.Run("UpdatedPublishesSlugOnlyWithoutCount", func(t *testing.T) {
		fixture := "lens-f-380"
		lens := Lens{}
		assert.NoError(t, UnscopedDb().First(&lens, "id = ?", LensFixtures.Get(fixture).ID).Error)
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Save(LensFixtures.Pointer(fixture)).Error) })

		updated := event.Subscribe("lenses.updated")
		t.Cleanup(func() { event.Unsubscribe(updated) })
		count := event.Subscribe("count.lenses")
		t.Cleanup(func() { event.Unsubscribe(count) })

		assert.NoError(t, lens.UpdateMakeModel("Sigma", "85mm F1.4"))
		// The slug must be preserved across a Make/Model rename so the published identity is stable.
		assert.Equal(t, LensFixtures.Get(fixture).LensSlug, lens.LensSlug)

		select {
		case msg := <-updated.Receiver:
			assert.Equal(t, "lenses.updated", msg.Name)
			slugs, ok := msg.Fields["entities"].([]string)
			assert.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
			assert.Equal(t, []string{lens.LensSlug}, slugs)
		case <-time.After(2 * time.Second):
			t.Fatal("expected one lenses.updated event")
		}
		// An update does not change the lens count, so no count event is published.
		select {
		case msg := <-count.Receiver:
			t.Fatalf("unexpected %s event on lens update", msg.Name)
		case <-time.After(200 * time.Millisecond):
		}
	})
}

func TestLens_SaveForm(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		lens := Lens{}
		assert.NoError(t, UnscopedDb().First(&lens, "id = ?", LensFixtures.Get("lens-f-380").ID).Error)
		defer assert.NoError(t, UnscopedDb().Save(LensFixtures.Pointer("lens-f-380")).Error)
		err := lens.SaveForm(&form.Lens{LensMake: "Sigma", LensModel: "85mm F1.4"})
		assert.NoError(t, err)
		assert.Equal(t, CameraMakes["Sigma"], lens.LensMake) // NewLens normalizes the make.
		assert.Equal(t, "85mm F1.4", lens.LensModel)
		assert.Equal(t, "lens-f-380", lens.LensSlug) // Slug is preserved across renames.
	})
	t.Run("NilForm", func(t *testing.T) {
		lens := &Lens{ID: LensFixtures.Get("lens-f-380").ID}
		assert.Error(t, lens.SaveForm(nil))
	})
	t.Run("EmptyMake", func(t *testing.T) {
		lens := &Lens{ID: LensFixtures.Get("lens-f-380").ID}
		assert.Error(t, lens.SaveForm(&form.Lens{LensMake: "", LensModel: "85mm F1.4"}))
	})
}

func TestAddLens(t *testing.T) {
	// removeLens deletes a test lens so that each case starts from a clean state.
	removeLens := func(slug string) {
		lensCache.Delete(slug)
		assert.NoError(t, UnscopedDb().Delete(&Lens{}, "lens_slug = ?", slug).Error)
	}

	t.Run("Created", func(t *testing.T) {
		slug := NewLens("Helios", "44-2 58mm f/2").LensSlug
		removeLens(slug)
		t.Cleanup(func() { removeLens(slug) })

		result, created, err := AddLens("  Helios ", " 44-2 58mm f/2  ")

		assert.NoError(t, err)
		assert.True(t, created)

		if result == nil {
			t.Fatal("result must not be nil")
		}

		assert.NotZero(t, result.ID)
		assert.Equal(t, slug, result.LensSlug)
		assert.Equal(t, "Helios 44-2 58mm f/2", result.LensName)
		assert.Equal(t, SrcManual, result.LensSrc)

		// The source must be persisted, not only set on the returned struct.
		found := Lens{}
		assert.NoError(t, Db().First(&found, "id = ?", result.ID).Error)
		assert.Equal(t, SrcManual, found.LensSrc)

		// Adding the same lens again reports the existing record instead of a duplicate.
		again, created, err := AddLens("Helios", "44-2 58mm f/2")
		assert.NoError(t, err)
		assert.False(t, created)
		assert.Equal(t, result.ID, again.ID)
	})
	t.Run("ExistingBySlug", func(t *testing.T) {
		fixture := LensFixtures.Get("4.15mm-f/2.2")
		t.Cleanup(func() {
			FlushLensCache()
			assert.NoError(t, UnscopedDb().Save(LensFixtures.Pointer("4.15mm-f/2.2")).Error)
		})

		result, created, err := AddLens(fixture.LensMake, fixture.LensModel)

		assert.NoError(t, err)
		assert.False(t, created)
		assert.Equal(t, fixture.ID, result.ID)
		assert.Equal(t, SrcManual, result.LensSrc)

		found := Lens{}
		assert.NoError(t, Db().First(&found, "id = ?", fixture.ID).Error)
		assert.Equal(t, SrcManual, found.LensSrc)
	})
	t.Run("ExistingRenamed", func(t *testing.T) {
		fixture := LensFixtures.Get("4.15mm-f/2.2")
		t.Cleanup(func() {
			FlushLensCache()
			assert.NoError(t, UnscopedDb().Save(LensFixtures.Pointer("4.15mm-f/2.2")).Error)
		})

		// A renamed record keeps its slug, so it can only be found by make and model.
		renamed := Lens{}
		assert.NoError(t, Db().First(&renamed, "id = ?", fixture.ID).Error)
		assert.NoError(t, renamed.UpdateMakeModel("Zeiss", "Planar 50mm f/1.4"))
		assert.Equal(t, fixture.LensSlug, renamed.LensSlug)
		assert.NotEqual(t, NewLens("Zeiss", "Planar 50mm f/1.4").LensSlug, renamed.LensSlug)

		result, created, err := AddLens("Zeiss", "Planar 50mm f/1.4")

		assert.NoError(t, err)
		assert.False(t, created)
		assert.Equal(t, fixture.ID, result.ID)
	})
	t.Run("EmptyMake", func(t *testing.T) {
		result, created, err := AddLens("  ", "44-2 58mm f/2")
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.ErrorContains(t, err, "make and model must not be empty")
		assert.False(t, created)
		assert.Nil(t, result)
	})
	t.Run("EmptyModel", func(t *testing.T) {
		result, created, err := AddLens("Helios", "")
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.ErrorContains(t, err, "make and model must not be empty")
		assert.False(t, created)
		assert.Nil(t, result)
	})
	t.Run("ModelSameAsMake", func(t *testing.T) {
		result, created, err := AddLens("Zenit", "Zenit")
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.ErrorContains(t, err, "model must not be empty after removing the make")
		assert.False(t, created)
		assert.Nil(t, result)
	})
	t.Run("UnknownSlug", func(t *testing.T) {
		unknown := Lens{}
		assert.NoError(t, Db().First(&unknown, "id = ?", UnknownLens.ID).Error)

		// These inputs normalize to the slug of the shared placeholder, which must stay untouched.
		for _, input := range [][2]string{{"ZZ", "."}, {"!", "zz"}, {"-", "zz"}} {
			result, created, err := AddLens(input[0], input[1])
			assert.ErrorIs(t, err, ErrInvalidValue, "%q", input)
			assert.False(t, created)
			assert.Nil(t, result)
		}

		found := Lens{}
		assert.NoError(t, Db().First(&found, "id = ?", UnknownLens.ID).Error)
		assert.Equal(t, unknown.LensSrc, found.LensSrc)
	})
	t.Run("ExistingPurged", func(t *testing.T) {
		slug := NewLens("Zenit", "TTL").LensSlug
		removeLens(slug)
		t.Cleanup(func() { removeLens(slug) })

		// A discovered orphan that is purged between the lookup and marking it is created again.
		orphan := FirstOrCreateLens(NewLens("Zenit", "TTL"))
		assert.NotZero(t, orphan.ID)
		removeLens(slug)
		assert.ErrorIs(t, orphan.markManual(), gorm.ErrRecordNotFound)

		result, created, err := AddLens("Zenit", "TTL")
		assert.NoError(t, err)
		assert.True(t, created)
		assert.Equal(t, SrcManual, result.LensSrc)
	})
}

func TestLens_markManual(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fixture := LensFixtures.Get("lens-f-380")
		t.Cleanup(func() {
			FlushLensCache()
			assert.NoError(t, UnscopedDb().Save(LensFixtures.Pointer("lens-f-380")).Error)
		})

		m := Lens{}
		assert.NoError(t, Db().First(&m, "id = ?", fixture.ID).Error)
		assert.Empty(t, m.LensSrc)
		assert.NoError(t, m.markManual())
		assert.Equal(t, SrcManual, m.LensSrc)

		found := Lens{}
		assert.NoError(t, Db().First(&found, "id = ?", fixture.ID).Error)
		assert.Equal(t, SrcManual, found.LensSrc)

		// Marking it again is a no-op.
		assert.NoError(t, m.markManual())
	})
	t.Run("NoPrimaryKey", func(t *testing.T) {
		m := Lens{LensSlug: "no-primary-key"}
		assert.Error(t, m.markManual())
		assert.Empty(t, m.LensSrc)
	})
}

func TestLens_UpdateMakeModelUnknown(t *testing.T) {
	m := UnknownLens
	assert.NotZero(t, m.ID)
	assert.EqualError(t, m.UpdateMakeModel("Helios", "44-2 58mm f/2"), "unknown lens cannot be changed")
	assert.Equal(t, UnknownLens.LensName, m.LensName)
}

// useLens temporarily assigns the first photo to the lens and returns the photo ID.
func useLens(t *testing.T, m *Lens) uint {
	t.Helper()

	photo := Photo{}
	assert.NoError(t, UnscopedDb().Order("id").First(&photo).Error)
	t.Cleanup(func() {
		assert.NoError(t, UnscopedDb().Model(&Photo{}).Where("id = ?", photo.ID).UpdateColumn("lens_id", photo.LensID).Error)
	})
	assert.NoError(t, UnscopedDb().Model(&Photo{}).Where("id = ?", photo.ID).UpdateColumn("lens_id", m.ID).Error)

	return photo.ID
}

// addLens adds a lens for testing and removes it again when the test is done.
func addLens(t *testing.T, makeName, modelName string) *Lens {
	t.Helper()

	slug := NewLens(makeName, modelName).LensSlug
	remove := func() {
		lensCache.Delete(slug)
		assert.NoError(t, UnscopedDb().Delete(&Lens{}, "lens_slug = ?", slug).Error)
	}
	remove()
	t.Cleanup(remove)

	m, created, err := AddLens(makeName, modelName)
	assert.NoError(t, err)
	assert.True(t, created)

	if m == nil {
		t.Fatal("lens must not be nil")
	}

	return m
}

func TestLens_PhotoCount(t *testing.T) {
	t.Run("Unused", func(t *testing.T) {
		m := addLens(t, "Jupiter", "9 85mm f/2")
		count, err := m.PhotoCount()
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
	t.Run("Used", func(t *testing.T) {
		m := addLens(t, "Jupiter", "8 50mm f/2")
		useLens(t, m)
		count, err := m.PhotoCount()
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
	t.Run("EmptyID", func(t *testing.T) {
		_, err := (&Lens{}).PhotoCount()
		assert.Error(t, err)
	})
}

func TestLens_Delete(t *testing.T) {
	exists := func(t *testing.T, id uint) bool {
		var count int
		assert.NoError(t, UnscopedDb().Model(&Lens{}).Where("id = ?", id).Count(&count).Error)
		return count > 0
	}

	t.Run("Unused", func(t *testing.T) {
		m := addLens(t, "Jupiter", "9 85mm f/2")

		deleted := event.Subscribe("lenses.deleted")
		t.Cleanup(func() { event.Unsubscribe(deleted) })

		reassigned, err := m.Delete(false)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), reassigned)
		assert.False(t, exists(t, m.ID))

		_, cached := lensCache.Get(m.LensSlug)
		assert.False(t, cached)

		select {
		case msg := <-deleted.Receiver:
			assert.Equal(t, []string{m.LensSlug}, msg.Fields["entities"])
		case <-time.After(2 * time.Second):
			t.Fatal("expected one lenses.deleted event")
		}
	})
	t.Run("InUse", func(t *testing.T) {
		m := addLens(t, "Jupiter", "8 50mm f/2")
		photoID := useLens(t, m)

		reassigned, err := m.Delete(false)
		assert.ErrorIs(t, err, ErrInUse)
		assert.Equal(t, int64(0), reassigned)
		assert.True(t, exists(t, m.ID))

		photo := Photo{}
		assert.NoError(t, UnscopedDb().First(&photo, "id = ?", photoID).Error)
		assert.Equal(t, m.ID, photo.LensID)
	})
	t.Run("Reassign", func(t *testing.T) {
		m := addLens(t, "Jupiter", "11 135mm f/4")
		photoID := useLens(t, m)

		reassigned, err := m.Delete(true)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), reassigned)
		assert.False(t, exists(t, m.ID))

		photo := Photo{}
		assert.NoError(t, UnscopedDb().First(&photo, "id = ?", photoID).Error)
		assert.Equal(t, UnknownLens.ID, photo.LensID)
	})
	t.Run("Unknown", func(t *testing.T) {
		m := UnknownLens
		_, err := m.Delete(true)
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.True(t, exists(t, UnknownLens.ID))
	})
	t.Run("NotFound", func(t *testing.T) {
		m := Lens{ID: 999999999, LensSlug: "not-found"}
		_, err := m.Delete(false)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("EmptyID", func(t *testing.T) {
		_, err := (&Lens{LensSlug: "empty-id"}).Delete(false)
		assert.Error(t, err)
	})
}

func TestFindLensesByMakeModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := addLens(t, "Jupiter", "3 50mm f/1.5")
		found := FindLensesByMakeModel("  Jupiter ", "Jupiter 3 50mm f/1.5")

		if assert.Len(t, found, 1) {
			assert.Equal(t, m.ID, found[0].ID)
		}
	})
	t.Run("Renamed", func(t *testing.T) {
		fixture := LensFixtures.Get("4.15mm-f/2.2")
		t.Cleanup(func() {
			FlushLensCache()
			assert.NoError(t, UnscopedDb().Save(LensFixtures.Pointer("4.15mm-f/2.2")).Error)
		})

		renamed := Lens{}
		assert.NoError(t, Db().First(&renamed, "id = ?", fixture.ID).Error)
		assert.NoError(t, renamed.UpdateMakeModel("Zeiss", "Planar 50mm f/1.4"))

		// The old name still matches the slug, but must no longer find the renamed record.
		assert.Empty(t, FindLensesByMakeModel(fixture.LensMake, fixture.LensModel))

		if found := FindLensesByMakeModel("Zeiss", "Planar 50mm f/1.4"); assert.Len(t, found, 1) {
			assert.Equal(t, fixture.ID, found[0].ID)
		}
	})
	t.Run("AsStored", func(t *testing.T) {
		// A record saved without normalizing its make is selected as stored, not its normalized twin.
		normalized := addLens(t, "Pentax", "37A 135mm f/3.5")
		assert.Equal(t, "PENTAX", normalized.LensMake)

		stored := Lens{LensSlug: "as-stored-lens-test", LensName: "Pentax 37A 135mm f/3.5", LensMake: "Pentax", LensModel: "37A 135mm f/3.5"}
		assert.NoError(t, UnscopedDb().Create(&stored).Error)
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&Lens{}, "id = ?", stored.ID).Error) })

		if found := FindLensesByMakeModel("Pentax", "37A 135mm f/3.5"); assert.Len(t, found, 1) {
			assert.Equal(t, stored.ID, found[0].ID)
		}

		if found := FindLensesByMakeModel("PENTAX", "37A 135mm f/3.5"); assert.Len(t, found, 1) {
			assert.Equal(t, normalized.ID, found[0].ID)
		}
	})
	t.Run("Duplicates", func(t *testing.T) {
		first := addLens(t, "Jupiter", "12 35mm f/2.8")

		second := Lens{LensSlug: "duplicate-lens-test", LensName: first.LensName, LensMake: first.LensMake, LensModel: first.LensModel}
		assert.NoError(t, UnscopedDb().Create(&second).Error)
		t.Cleanup(func() { assert.NoError(t, UnscopedDb().Delete(&Lens{}, "id = ?", second.ID).Error) })

		if found := FindLensesByMakeModel("Jupiter", "12 35mm f/2.8"); assert.Len(t, found, 2) {
			assert.Equal(t, first.ID, found[0].ID)
			assert.Equal(t, second.ID, found[1].ID)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		assert.Empty(t, FindLensesByMakeModel("Jupiter", "Does Not Exist"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Empty(t, FindLensesByMakeModel("", ""))
		assert.Empty(t, FindLensesByMakeModel("ZZ", "."))
		assert.Empty(t, FindLensesByMakeModel("", "Unknown"))
	})
	t.Run("ModelSameAsMake", func(t *testing.T) {
		assert.Empty(t, FindLensesByMakeModel("Jupiter", "Jupiter"))
	})
}

func TestFindLensesByMakeModelExact(t *testing.T) {
	t.Run("Exact", func(t *testing.T) {
		m := addLens(t, "Jupiter", "3 50mm f/1.5")

		if found := findLensesByMakeModel(m.LensMake, m.LensModel); assert.Len(t, found, 1) {
			assert.Equal(t, m.ID, found[0].ID)
		}
	})
	t.Run("CaseDiffers", func(t *testing.T) {
		m := addLens(t, "Jupiter", "3 50mm f/1.5")

		// MariaDB collations match this in SQL, so the result must be filtered in Go.
		assert.Empty(t, findLensesByMakeModel(strings.ToUpper(m.LensMake), m.LensModel))
	})
	t.Run("Placeholder", func(t *testing.T) {
		assert.Empty(t, findLensesByMakeModel(UnknownLens.LensMake, UnknownLens.LensModel))
	})
}

func TestUnknownLensID(t *testing.T) {
	placeholder := Lens{}
	assert.NoError(t, UnscopedDb().First(&placeholder, "lens_slug = ?", UnknownID).Error)

	t.Run("Initialized", func(t *testing.T) {
		id, err := unknownLensID()
		assert.NoError(t, err)
		assert.Equal(t, placeholder.ID, id)
	})
	t.Run("Uninitialized", func(t *testing.T) {
		// The CLI does not initialize the placeholder on startup, so its ID must be read from the database.
		prev := UnknownLens
		t.Cleanup(func() { UnknownLens = prev })
		UnknownLens.ID = 0
		FlushLensCache()

		id, err := unknownLensID()
		assert.NoError(t, err)
		assert.Equal(t, placeholder.ID, id)
	})
}

func TestFindExistingLens(t *testing.T) {
	t.Run("BySlug", func(t *testing.T) {
		fixture := LensFixtures.Get("4.15mm-f/2.2")
		found := findExistingLens(NewLens(fixture.LensMake, fixture.LensModel))

		if found == nil {
			t.Fatal("lens must be found")
		}

		assert.Equal(t, fixture.ID, found.ID)
	})
	t.Run("ByMakeModel", func(t *testing.T) {
		fixture := LensFixtures.Get("4.15mm-f/2.2")
		t.Cleanup(func() {
			FlushLensCache()
			assert.NoError(t, UnscopedDb().Save(LensFixtures.Pointer("4.15mm-f/2.2")).Error)
		})

		renamed := Lens{}
		assert.NoError(t, Db().First(&renamed, "id = ?", fixture.ID).Error)
		assert.NoError(t, renamed.UpdateMakeModel("Zeiss", "Planar 50mm f/1.4"))

		found := findExistingLens(NewLens("Zeiss", "Planar 50mm f/1.4"))

		if found == nil {
			t.Fatal("lens must be found")
		}

		assert.Equal(t, fixture.ID, found.ID)
	})
	t.Run("None", func(t *testing.T) {
		assert.Nil(t, findExistingLens(NewLens("Jupiter", "Does Not Exist")))
	})
}

func TestLens_DeleteEdgeCases(t *testing.T) {
	exists := func(t *testing.T, id uint) bool {
		var count int
		assert.NoError(t, UnscopedDb().Model(&Lens{}).Where("id = ?", id).Count(&count).Error)
		return count > 0
	}

	placeholder := Lens{}
	assert.NoError(t, UnscopedDb().First(&placeholder, "lens_slug = ?", UnknownID).Error)

	t.Run("OnlyTheRecord", func(t *testing.T) {
		m := addLens(t, "Jupiter", "3 50mm f/1.5")
		other := addLens(t, "Jupiter", "37A 135mm f/3.5")

		_, err := m.Delete(false)
		assert.NoError(t, err)
		assert.False(t, exists(t, m.ID))
		assert.True(t, exists(t, other.ID))
	})
	t.Run("SoftDeletedPhoto", func(t *testing.T) {
		m := addLens(t, "Jupiter", "12 35mm f/2.8")
		photoID := useLens(t, m)

		// Archived and deleted pictures still reference the lens and must be counted and reassigned.
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
		assert.Equal(t, placeholder.ID, photo.LensID)
	})
	t.Run("UninitializedPlaceholder", func(t *testing.T) {
		m := addLens(t, "Jupiter", "21M 200mm f/4")
		photoID := useLens(t, m)

		// The CLI does not initialize the placeholder, so reassigning must not write ID 0.
		prev := UnknownLens
		t.Cleanup(func() { UnknownLens = prev })
		UnknownLens.ID = 0
		FlushLensCache()

		reassigned, err := m.Delete(true)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), reassigned)

		photo := Photo{}
		assert.NoError(t, UnscopedDb().First(&photo, "id = ?", photoID).Error)
		assert.Equal(t, placeholder.ID, photo.LensID)
	})
	t.Run("PlaceholderID", func(t *testing.T) {
		m := Lens{ID: placeholder.ID, LensSlug: "not-the-placeholder-slug"}
		_, err := m.Delete(true)
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.True(t, exists(t, placeholder.ID))
	})
	t.Run("PlaceholderSlug", func(t *testing.T) {
		m := addLens(t, "Jupiter", "6 180mm f/2.8")
		m.LensSlug = UnknownID
		_, err := m.Delete(true)
		assert.ErrorIs(t, err, ErrInvalidValue)
		assert.True(t, exists(t, m.ID))
	})
}
