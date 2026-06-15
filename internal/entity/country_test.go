package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/event"
)

func TestNewCountry(t *testing.T) {
	t.Run("UnknownCountry", func(t *testing.T) {
		country := NewCountry("", "")

		assert.Equal(t, &UnknownCountry, country)
	})
	t.Run("UnitedStates", func(t *testing.T) {
		country := NewCountry("us", "United States")

		assert.Equal(t, "United States", country.CountryName)
		assert.Equal(t, "united-states", country.CountrySlug)
	})
	t.Run("Germany", func(t *testing.T) {
		country := NewCountry("de", "Germany")

		assert.Equal(t, "Germany", country.CountryName)
		assert.Equal(t, "germany", country.CountrySlug)
	})
}

func TestFirstOrCreateCountry(t *testing.T) {
	t.Run("Es", func(t *testing.T) {
		country := NewCountry("es", "spain")
		country = FirstOrCreateCountry(country)
		if country == nil {
			t.Fatal("country must not be nil")
		}
	})
	t.Run("De", func(t *testing.T) {
		country := &Country{ID: "de"}
		r := FirstOrCreateCountry(country)
		if r == nil {
			t.Fatal("country must not be nil")
		}
	})
}

// TestCountry_EntityEvents pins the countries content-channel payload to the UID-only
// invariant: countries.created carries a []string of stable country codes, never entity fields.
func TestCountry_EntityEvents(t *testing.T) {
	t.Run("CreatedPublishesCodeOnly", func(t *testing.T) {
		const code = "qq" // User-assigned ISO 3166 range; not a real fixture country.

		// Force the create branch to fire regardless of prior runs, -count>1, or cache state.
		removeTestCountry := func() {
			countryCache.Delete(code)
			assert.NoError(t, UnscopedDb().Delete(&Country{}, "id = ?", code).Error)
		}
		removeTestCountry()
		t.Cleanup(removeTestCountry)

		sub := event.Subscribe("countries.created")
		t.Cleanup(func() { event.Unsubscribe(sub) })

		country := FirstOrCreateCountry(NewCountry(code, "Event Test Country"))

		if country == nil {
			t.Fatal("country must not be nil")
		}

		assert.Equal(t, code, country.ID)

		select {
		case msg := <-sub.Receiver:
			assert.Equal(t, "countries.created", msg.Name)
			ids, ok := msg.Fields["entities"].([]string)
			assert.True(t, ok, "entities payload should be []string, got %T", msg.Fields["entities"])
			assert.Equal(t, []string{code}, ids)
		case <-time.After(2 * time.Second):
			t.Fatal("expected one countries.created event")
		}
	})
}

func TestCountry_Name(t *testing.T) {
	country := NewCountry("xy", "Neverland")
	assert.Equal(t, "Neverland", country.Name())
}

func TestCountry_Code(t *testing.T) {
	country := NewCountry("xy", "Neverland")
	assert.Equal(t, "xy", country.Code())
}
