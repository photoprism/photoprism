package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGeoResult_Lat(t *testing.T) {
	geo := GeoResult{
		ID:            "123",
		PhotoLat:      7.775,
		PhotoLng:      8.775,
		PhotoUID:      "",
		PhotoTitle:    "",
		PhotoFavorite: false,
		FileHash:      "",
		FileWidth:     0,
		FileHeight:    0,
		TakenAt:       time.Time{},
	}
	assert.Equal(t, 7.775000095367432, geo.Lat())
}

func TestGeoResult_Lng(t *testing.T) {
	geo := GeoResult{
		ID:            "123",
		PhotoLat:      7.775,
		PhotoLng:      8.775,
		PhotoUID:      "",
		PhotoTitle:    "",
		PhotoFavorite: false,
		FileHash:      "",
		FileWidth:     0,
		FileHeight:    0,
		TakenAt:       time.Time{},
	}
	assert.Equal(t, 8.774999618530273, geo.Lng())
}
