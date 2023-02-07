package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUnknownPlace(t *testing.T) {
	r := FirstOrCreatePlace(&UnknownPlace)
	assert.True(t, r.Unknown())
}

func TestFindPlace(t *testing.T) {
	t.Run("Holiday Park", func(t *testing.T) {
		r := FindPlace("de:HFqPHxa2Hsol")

		if r == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, "de", r.PlaceCountry)
	})
	t.Run("Mexico", func(t *testing.T) {
		r := FindPlace("mx:VvfNBpFegSCr")

		if r == nil {
			t.Fatal("result should not be nil")
		}
		assert.Equal(t, "mx", r.PlaceCountry)
	})
	t.Run("KwaDukuza", func(t *testing.T) {
		r := FindPlace("za:Rc1K7dTWRzBD")

		if r == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, "za", r.PlaceCountry)
	})
	t.Run("not matching", func(t *testing.T) {
		r := FindPlace("111")

		if r != nil {
			t.Fatal("result should be nil")
		}
	})
	t.Run("not matching empty label", func(t *testing.T) {
		r := FindPlace("111")

		if r != nil {
			t.Fatal("result should be nil")
		}
	})
}

func TestFirstOrCreatePlace(t *testing.T) {
	t.Run("existing place", func(t *testing.T) {
		m := PlaceFixtures.Pointer("zinkwazi")
		r := FirstOrCreatePlace(m)
		assert.Equal(t, "KwaDukuza, KwaZulu-Natal, South Africa", r.PlaceLabel)
	})
	t.Run("ID empty", func(t *testing.T) {
		p := &Place{ID: ""}
		assert.Nil(t, FirstOrCreatePlace(p))
	})
	t.Run("PlaceLabel empty", func(t *testing.T) {
		p := &Place{ID: "abcde44", PlaceLabel: ""}
		assert.Nil(t, FirstOrCreatePlace(p))
	})
}

func TestPlace_LongCity(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := Place{PlaceCity: "veryveryveryverylongcity"}
		assert.True(t, p.LongCity())
	})
	t.Run("false", func(t *testing.T) {
		p := Place{PlaceCity: "short"}
		assert.False(t, p.LongCity())
	})
}

func TestPlace_NoCity(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := Place{PlaceCity: ""}
		assert.True(t, p.NoCity())
	})
	t.Run("false", func(t *testing.T) {
		p := Place{PlaceCity: "short"}
		assert.False(t, p.NoCity())
	})
}

func TestPlace_CityContains(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := Place{PlaceCity: "Munich"}
		assert.True(t, p.CityContains("Munich"))
	})
	t.Run("false", func(t *testing.T) {
		p := Place{PlaceCity: "short"}
		assert.False(t, p.CityContains("ich"))
	})
}
