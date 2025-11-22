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
	t.Run("HolidayPark", func(t *testing.T) {
		r := FindPlace("de:HFqPHxa2Hsol")

		if r == nil {
			t.Fatal("result must not be nil")
		}

		assert.Equal(t, "de", r.PlaceCountry)
	})
	t.Run("Mexico", func(t *testing.T) {
		r := FindPlace("mx:VvfNBpFegSCr")

		if r == nil {
			t.Fatal("result must not be nil")
		}
		assert.Equal(t, "mx", r.PlaceCountry)
	})
	t.Run("KwaDukuza", func(t *testing.T) {
		r := FindPlace("za:Rc1K7dTWRzBD")

		if r == nil {
			t.Fatal("result must not be nil")
		}

		assert.Equal(t, "za", r.PlaceCountry)
	})
	t.Run("NotMatching", func(t *testing.T) {
		r := FindPlace("111")

		if r != nil {
			t.Fatal("result should be nil")
		}
	})
	t.Run("NotMatchingEmptyLabel", func(t *testing.T) {
		r := FindPlace("111")

		if r != nil {
			t.Fatal("result should be nil")
		}
	})
}

func TestFirstOrCreatePlace(t *testing.T) {
	t.Run("ExistingPlace", func(t *testing.T) {
		m := PlaceFixtures.Pointer("zinkwazi")
		r := FirstOrCreatePlace(m)
		assert.Equal(t, "KwaDukuza, KwaZulu-Natal, South Africa", r.PlaceLabel)
	})
	t.Run("IdEmpty", func(t *testing.T) {
		p := &Place{ID: ""}
		assert.Nil(t, FirstOrCreatePlace(p))
	})
	t.Run("PlaceLabelEmpty", func(t *testing.T) {
		p := &Place{ID: "abcde44", PlaceLabel: ""}
		assert.Nil(t, FirstOrCreatePlace(p))
	})
}

func TestPlace_LongCity(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		p := Place{PlaceCity: "veryveryveryverylongcity"}
		assert.True(t, p.LongCity())
	})
	t.Run("False", func(t *testing.T) {
		p := Place{PlaceCity: "short"}
		assert.False(t, p.LongCity())
	})
}

func TestPlace_NoCity(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		p := Place{PlaceCity: ""}
		assert.True(t, p.NoCity())
	})
	t.Run("False", func(t *testing.T) {
		p := Place{PlaceCity: "short"}
		assert.False(t, p.NoCity())
	})
}

func TestPlace_CityContains(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		p := Place{PlaceCity: "Munich"}
		assert.True(t, p.CityContains("Munich"))
	})
	t.Run("False", func(t *testing.T) {
		p := Place{PlaceCity: "short"}
		assert.False(t, p.CityContains("ich"))
	})
}
