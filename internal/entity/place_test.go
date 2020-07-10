package entity

import (
	"testing"

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
	t.Run("not matching empty label", func(t *testing.T) {
		r := FindPlace("111", "")

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
			CreatedAt:   Timestamp(),
			UpdatedAt:   Timestamp(),
			New:         false,
		}
		err := place.Find()
		assert.EqualError(t, err, "record not found")
	})
}

func TestFirstOrCreatePlace(t *testing.T) {
	t.Run("existing place", func(t *testing.T) {
		m := PlaceFixtures.Pointer("zinkwazi")
		r := FirstOrCreatePlace(m)
		assert.Equal(t, "KwaDukuza, KwaZulu-Natal, South Africa", r.LocLabel)
	})
	t.Run("ID empty", func(t *testing.T) {
		p := &Place{ID: ""}
		assert.Nil(t, FirstOrCreatePlace(p))
	})
	t.Run("LocLabel empty", func(t *testing.T) {
		p := &Place{ID: "abcde44", LocLabel: ""}
		assert.Nil(t, FirstOrCreatePlace(p))
	})
}

func TestPlace_LongCity(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := Place{LocCity: "veryveryveryverylongcity"}
		assert.True(t, p.LongCity())
	})
	t.Run("false", func(t *testing.T) {
		p := Place{LocCity: "short"}
		assert.False(t, p.LongCity())
	})
}

func TestPlace_NoCity(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := Place{LocCity: ""}
		assert.True(t, p.NoCity())
	})
	t.Run("false", func(t *testing.T) {
		p := Place{LocCity: "short"}
		assert.False(t, p.NoCity())
	})
}

func TestPlace_CityContains(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := Place{LocCity: "Munich"}
		assert.True(t, p.CityContains("Munich"))
	})
	t.Run("false", func(t *testing.T) {
		p := Place{LocCity: "short"}
		assert.False(t, p.CityContains("ich"))
	})
}
