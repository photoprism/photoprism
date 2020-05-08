package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
		place := &Place{"1110", "test", "testCity", "", "", "", "", false, time.Now(), time.Now(), false}
		r := place.Find()
		assert.Equal(t, "record not found", r.Error())
	})
}

func TestPlace_FirstOrCreate(t *testing.T) {
	r := PlaceFixtureZinkwazi.FirstOrCreate()
	assert.Equal(t, "KwaDukuza, KwaZulu-Natal, South Africa", r.LocLabel)
}
