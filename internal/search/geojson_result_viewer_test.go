package search

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestGeoResults_ViewerJSON(t *testing.T) {
	taken := time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC).UTC().Round(time.Second)
	items := GeoResults{
		GeoResult{
			ID:               "1",
			PhotoLat:         7.775,
			PhotoLng:         8.775,
			PhotoUID:         "p1",
			PhotoTitle:       "Title 1",
			PhotoDescription: "Description 1",
			PhotoFavorite:    false,
			PhotoType:        entity.TypeVideo,
			FileHash:         "d2b4a5d18276f96f1b5a1bf17fd82d6fab3807f2",
			FileWidth:        1920,
			FileHeight:       1080,
			TakenAt:          taken,
		},
		GeoResult{
			ID:               "2",
			PhotoLat:         1.775,
			PhotoLng:         -5.775,
			PhotoUID:         "p2",
			PhotoTitle:       "Title 2",
			PhotoDescription: "Description 2",
			PhotoFavorite:    true,
			PhotoType:        entity.TypeImage,
			FileHash:         "da639e836dfa9179e66c619499b0a5e592f72fc1",
			FileWidth:        3024,
			FileHeight:       3024,
			TakenAt:          taken,
		},
		GeoResult{
			ID:               "3",
			PhotoLat:         -1.775,
			PhotoLng:         100.775,
			PhotoUID:         "p3",
			PhotoTitle:       "Title 3",
			PhotoDescription: "Description 3",
			PhotoFavorite:    false,
			PhotoType:        entity.TypeRaw,
			FileHash:         "412fe4c157a82b636efebc5bc4bc4a15c321aad1",
			FileWidth:        5000,
			FileHeight:       10000,
			TakenAt:          taken,
		},
	}

	b, err := items.ViewerJSON("/content", "/api/v1", "preview-token", "download-token")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("result: %s", b)
}
