package entity

import (
	"testing"
	"time"

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
