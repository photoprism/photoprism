package search

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
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
		TakenAtLocal:  time.Time{},
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
		TakenAtLocal:  time.Time{},
	}
	assert.Equal(t, 8.774999618530273, geo.Lng())
}

func TestGeoResults_GeoJSON(t *testing.T) {
	taken := time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC).UTC().Truncate(time.Second)
	items := GeoResults{
		GeoResult{
			ID:               "1",
			PhotoLat:         7.775,
			PhotoLng:         8.775,
			PhotoUID:         "p1",
			PhotoTitle:       "Title 1",
			PhotoDescription: "Description 1",
			PhotoFavorite:    false,
			PhotoType:        entity.MediaVideo,
			FileHash:         "d2b4a5d18276f96f1b5a1bf17fd82d6fab3807f2",
			FileWidth:        1920,
			FileHeight:       1080,
			TakenAtLocal:     taken,
		},
		GeoResult{
			ID:               "2",
			PhotoLat:         1.775,
			PhotoLng:         -5.775,
			PhotoUID:         "p2",
			PhotoTitle:       "Title 2",
			PhotoDescription: "Description 2",
			PhotoFavorite:    true,
			PhotoType:        entity.MediaImage,
			FileHash:         "da639e836dfa9179e66c619499b0a5e592f72fc1",
			FileWidth:        3024,
			FileHeight:       3024,
			TakenAtLocal:     taken,
		},
		GeoResult{
			ID:               "3",
			PhotoLat:         -1.775,
			PhotoLng:         100.775,
			PhotoUID:         "p3",
			PhotoTitle:       "Title 3",
			PhotoDescription: "Description 3",
			PhotoFavorite:    false,
			PhotoType:        entity.MediaRaw,
			FileHash:         "412fe4c157a82b636efebc5bc4bc4a15c321aad1",
			FileWidth:        5000,
			FileHeight:       10000,
			TakenAtLocal:     taken,
		},
	}

	b, err := items.GeoJSON()

	if err != nil {
		t.Fatal(err)
	}

	expected := []byte("{\"type\":\"FeatureCollection\",\"bbox\":[-5.775000095367432,-1.774999976158142,100.7750015258789,7.775000095367432]")

	assert.Truef(t, bytes.Contains(b, expected), "GeoJSON not as expected")

	t.Logf("result: %s", b)
}
