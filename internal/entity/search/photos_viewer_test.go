package search

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestPhotoResults_ViewerJSON(t *testing.T) {
	result1 := Photo{
		ID:               111111,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
		DeletedAt:        &time.Time{},
		TakenAt:          time.Time{},
		TakenAtLocal:     time.Time{},
		TakenSrc:         "",
		TimeZone:         "",
		PhotoUID:         "123",
		PhotoPath:        "",
		PhotoName:        "",
		PhotoTitle:       "Photo1",
		PhotoYear:        0,
		PhotoMonth:       0,
		PhotoCountry:     "",
		PhotoFavorite:    false,
		PhotoPrivate:     false,
		PhotoLat:         0,
		PhotoLng:         0,
		PhotoAltitude:    0,
		PhotoIso:         0,
		PhotoFocalLength: 0,
		PhotoFNumber:     0,
		PhotoExposure:    "",
		PhotoQuality:     0,
		PhotoResolution:  0,
		Merged:           false,
		CameraID:         0,
		CameraModel:      "",
		CameraMake:       "",
		LensID:           0,
		LensModel:        "",
		LensMake:         "",
		CellID:           "",
		PlaceID:          "",
		PlaceLabel:       "",
		PlaceCity:        "",
		PlaceState:       "",
		PlaceCountry:     "",
		FileID:           0,
		FileUID:          "",
		FilePrimary:      false,
		FileMissing:      false,
		FileName:         "",
		FileHash:         "",
		FileType:         "",
		FileMime:         "",
		FileWidth:        0,
		FileHeight:       0,
		FileOrientation:  0,
		FileAspectRatio:  0,
		FileColors:       "",
		FileChroma:       0,
		FileLuminance:    "",
		FileDiff:         0,
		Files:            nil,
	}

	result2 := Photo{
		ID:               22222,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
		DeletedAt:        &time.Time{},
		TakenAt:          time.Time{},
		TakenAtLocal:     time.Time{},
		TakenSrc:         "",
		TimeZone:         "",
		PhotoUID:         "456",
		PhotoPath:        "",
		PhotoName:        "",
		PhotoTitle:       "Photo2",
		PhotoYear:        0,
		PhotoMonth:       0,
		PhotoCountry:     "",
		PhotoFavorite:    false,
		PhotoPrivate:     false,
		PhotoLat:         0,
		PhotoLng:         0,
		PhotoAltitude:    0,
		PhotoIso:         0,
		PhotoFocalLength: 0,
		PhotoFNumber:     0,
		PhotoExposure:    "",
		PhotoQuality:     0,
		PhotoResolution:  0,
		Merged:           false,
		CameraID:         0,
		CameraModel:      "",
		CameraMake:       "",
		LensID:           0,
		LensModel:        "",
		LensMake:         "",
		CellID:           "",
		PlaceID:          "",
		PlaceLabel:       "",
		PlaceCity:        "",
		PlaceState:       "",
		PlaceCountry:     "",
		FileID:           0,
		FileUID:          "",
		FilePrimary:      false,
		FileMissing:      false,
		FileName:         "",
		FileHash:         "",
		FileType:         "",
		FileMime:         "",
		FileWidth:        0,
		FileHeight:       0,
		FileOrientation:  0,
		FileAspectRatio:  0,
		FileColors:       "",
		FileChroma:       0,
		FileLuminance:    "",
		FileDiff:         0,
		Files:            nil,
	}

	results := PhotoResults{result1, result2}

	b, err := results.ViewerJSON("/content", "/api/v1", "preview-token", "download-token")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("result: %s", b)
}

func TestGeoResults_ViewerJSON(t *testing.T) {
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

	b, err := items.ViewerJSON("/content", "/api/v1", "preview-token", "download-token")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("result: %s", b)
}
