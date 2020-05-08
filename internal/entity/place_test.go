package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateUnknownPlace(t *testing.T) {
	r := UnknownPlace.FirstOrCreate()
	assert.True(t, r.Unknown())
}

func TestFindPlaceByLabel(t *testing.T) {
	t.Run("find by id", func(t *testing.T) {
		r := FindPlaceByLabel("1000000", "")
		assert.Equal(t, "mx", r.LocCountry)
	})
	t.Run("find by label", func(t *testing.T) {
		r := FindPlaceByLabel("", "KwaDukuza, KwaZulu-Natal, South Africa")
		assert.Equal(t, "za", r.LocCountry)
	})
	t.Run("not matching", func(t *testing.T) {
		r := FindPlaceByLabel("111", "xxx")
		assert.Nil(t, r)
	})
}

func TestPlace_Find(t *testing.T) {
	t.Run("record exists", func(t *testing.T) {
		r := PlaceFixtureTeotihuacan.Find()
		assert.Nil(t, r)
	})
	t.Run("record does not exist", func(t *testing.T) {
		place := &Place{
			ID:          "1110",
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
		r := place.Find()
		assert.Equal(t, "record not found", r.Error())
	})
}

func TestPlace_FirstOrCreate(t *testing.T) {
	r := PlaceFixtureZinkwazi.FirstOrCreate()
	assert.Equal(t, "KwaDukuza, KwaZulu-Natal, South Africa", r.LocLabel)
}
