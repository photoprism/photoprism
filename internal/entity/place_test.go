package entity

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/stretchr/testify/assert"
)

func TestCreateUnknownPlace(t *testing.T) {
	r := FirstOrCreatePlace(&UnknownPlace)
	assert.True(t, r.Unknown())
}

func TestFindPlaceByLabel(t *testing.T) {
	t.Run("find by id", func(t *testing.T) {
		r := FindPlace(s2.TokenPrefix+"1ef744d1e280", "")

		if r == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, "de", r.LocCountry)
	})
	t.Run("find by id", func(t *testing.T) {
		r := FindPlace(s2.TokenPrefix+"85d1ea7d3278", "")

		if r == nil {
			t.Fatal("result should not be nil")
		}
		assert.Equal(t, "mx", r.LocCountry)
	})
	t.Run("find by label", func(t *testing.T) {
		r := FindPlace("", "KwaDukuza, KwaZulu-Natal, South Africa")

		if r == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, "za", r.LocCountry)
	})
	t.Run("not matching", func(t *testing.T) {
		r := FindPlace("111", "xxx")

		if r != nil {
			t.Fatal("result should be nil")
		}
	})
}

func TestPlace_Find(t *testing.T) {
	t.Run("record exists", func(t *testing.T) {
		m := PlaceFixtures.Get("mexico")
		if err := m.Find(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("record does not exist", func(t *testing.T) {
		place := &Place{
			ID:          s2.TokenPrefix + "1110",
			LocLabel:    "test",
			LocCity:     "testCity",
			LocState:    "",
			LocCountry:  "",
			LocKeywords: "",
			LocNotes:    "",
			LocFavorite: false,
			PhotoCount:  0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			New:         false,
		}
		err := place.Find()
		assert.EqualError(t, err, "record not found")
	})
}

func TestFirstOrCreatePlace(t *testing.T) {
	m := PlaceFixtures.Pointer("zinkwazi")
	r := FirstOrCreatePlace(m)
	assert.Equal(t, "KwaDukuza, KwaZulu-Natal, South Africa", r.LocLabel)
}
