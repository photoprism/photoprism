package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestPhotoByID validates photo query behavior.
func TestPhotoByID(t *testing.T) {
	t.Run("PhotoFound", func(t *testing.T) {
		result, err := PhotoByID(1000000)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 2790, result.PhotoYear)
	})
	t.Run("NoPhotoFound", func(t *testing.T) {
		result, err := PhotoByID(99999)
		assert.Error(t, err, "record not found")
		t.Log(result)
	})
}

// TestPhotoByUID validates photo query behavior.
func TestPhotoByUID(t *testing.T) {
	t.Run("PhotoFound", func(t *testing.T) {
		result, err := PhotoByUID("ps6sg6be2lvl0y12")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Reunion", result.PhotoTitle)
	})
	t.Run("NoPhotoFound", func(t *testing.T) {
		result, err := PhotoByUID("99999")
		assert.Error(t, err, "record not found")
		t.Log(result)
	})
}

// TestPreloadPhotoByUID validates photo query behavior.
func TestPreloadPhotoByUID(t *testing.T) {
	t.Run("PhotoFound", func(t *testing.T) {
		result, err := PhotoPreloadByUID("ps6sg6be2lvl0y12")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Reunion", result.PhotoTitle)
	})
	t.Run("NoPhotoFound", func(t *testing.T) {
		result, err := PhotoPreloadByUID("99999")
		assert.Error(t, err, "record not found")
		t.Log(result)
	})
}

// TestPhotoPreloadByUIDs validates photo query behavior.
func TestPhotoPreloadByUIDs(t *testing.T) {
	t.Run("Multiple", func(t *testing.T) {
		uids := []string{"ps6sg6be2lvl0y12", "ps6sg6be2lvl0y25", "ps6sg6be2lvl0y12"}
		photos, err := PhotoPreloadByUIDs(uids)
		if err != nil {
			t.Fatal(err)
		}

		if len(photos) != 2 {
			t.Fatalf("expected two unique photos, got %d", len(photos))
		}

		photoMap := make(map[string]*entity.Photo, len(photos))
		for _, p := range photos {
			if p == nil {
				continue
			}
			photoMap[p.PhotoUID] = p
		}

		first := photoMap["ps6sg6be2lvl0y12"]
		if first == nil {
			t.Fatalf("expected photo ps6sg6be2lvl0y12 to be preloaded")
		}
		assert.Greater(t, len(first.Files), 0)
		assert.True(t, first.CameraID > 0)

		second := photoMap["ps6sg6be2lvl0y25"]
		if second == nil {
			t.Fatalf("expected photo ps6sg6be2lvl0y25 to be preloaded")
		}
		assert.Greater(t, len(second.Labels), 0)
	})

	t.Run("Empty", func(t *testing.T) {
		photos, err := PhotoPreloadByUIDs(nil)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
}

// TestMissingPhotos validates photo query behavior.
func TestMissingPhotos(t *testing.T) {
	result, err := MissingPhotos(15, 0)

	if err != nil {
		t.Fatal(err)
	}

	assert.LessOrEqual(t, 1, len(result))
}

// TestArchivedPhotos validates photo query behavior.
func TestArchivedPhotos(t *testing.T) {
	results, err := ArchivedPhotos(15, 0)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(results))

	if len(results) > 1 {
		result := results[0]
		assert.Equal(t, "image", result.PhotoType)
		assert.Equal(t, "ps6sg6be2lvl0y25", result.PhotoUID)
	}
}

// TestPhotosMetadataUpdate validates photo query behavior.
func TestPhotosMetadataUpdate(t *testing.T) {
	interval := entity.MetadataUpdateInterval
	result, err := PhotosMetadataUpdate(10, 0, time.Second, interval)

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, entity.Photos{}, result)
}

// TestOrphanPhotos validates photo query behavior.
func TestOrphanPhotos(t *testing.T) {
	result, err := OrphanPhotos()

	if err != nil {
		t.Fatal(err)
	}

	assert.IsType(t, entity.Photos{}, result)
}

// TODO How to verify?
// TestFixPrimaries validates photo query behavior.
func TestFixPrimaries(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		err := FixPrimaries()
		if err != nil {
			t.Fatal(err)
		}
	})
}

// TestFlagHiddenPhotos validates photo query behavior.
func TestFlagHiddenPhotos(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Set photo quality scores to -1 if files are missing.
		if err := FlagHiddenPhotos(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("SuccessWith1000", func(t *testing.T) {
		var checkedTime = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		// Load 1000 photos that need to be hidden
		for i := 0; i < 1000; i++ {
			newPhoto := entity.Photo{ //JPG, Geo from metadata, indexed
				//ID:               1000049,
				PhotoUID:         rnd.GenerateUID(entity.PhotoUID),
				TakenAt:          time.Date(2020, 11, 11, 9, 7, 18, 0, time.UTC),
				TakenAtLocal:     time.Date(2020, 11, 11, 9, 7, 18, 0, time.UTC),
				TakenSrc:         entity.SrcMeta,
				PhotoType:        "image",
				TypeSrc:          "",
				PhotoTitle:       "desk\"",
				TitleSrc:         entity.SrcManual,
				PhotoCaption:     "",
				CaptionSrc:       "",
				PhotoPath:        "2000\"/02\"",
				PhotoName:        "SuccessWith1000",
				OriginalName:     "",
				PhotoFavorite:    false,
				PhotoPrivate:     false,
				PhotoScan:        false,
				PhotoPanorama:    false,
				TimeZone:         "America/Mexico_City",
				PlaceSrc:         "meta",
				CellAccuracy:     0,
				PhotoAltitude:    3,
				PhotoLat:         48.519234,
				PhotoLng:         9.057997,
				PhotoCountry:     entity.CellFixtures.Pointer("caravan park").Place.CountryCode(),
				PhotoYear:        2020,
				PhotoMonth:       11,
				PhotoDay:         11,
				PhotoIso:         0,
				PhotoExposure:    "",
				PhotoFocalLength: 0,
				PhotoFNumber:     0,
				PhotoQuality:     5,
				PhotoResolution:  0,
				Camera:           entity.CameraFixtures.Pointer("canon-eos-6d"),
				CameraID:         entity.CameraFixtures.Pointer("canon-eos-6d").ID,
				CameraSerial:     "",
				CameraSrc:        "",
				Lens:             entity.LensFixtures.Pointer("lens-f-380"),
				LensID:           entity.LensFixtures.Pointer("lens-f-380").ID,
				Keywords:         []entity.Keyword{},
				Albums:           []entity.Album{},
				Files:            []entity.File{},
				Labels:           []entity.PhotoLabel{},
				CreatedAt:        time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				EditedAt:         nil,
				CheckedAt:        &checkedTime,
				DeletedAt:        nil,
				PhotoColor:       14,
				PhotoStack:       0,
				PhotoFaces:       0,
			}
			if err := Db().Create(&newPhoto).Error; err != nil {
				t.Fatal(err)
			}
		}
		// Set photo quality scores to -1 if files are missing.
		if err := FlagHiddenPhotos(); err != nil {
			t.Fatal(err)
		}

		var actual int64
		var expected int64 = 1000
		if err := Db().Model(&entity.Photo{}).Where("photo_name = ? AND photo_quality = ?", "SuccessWith1000", -1).Count(&actual).Error; err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, expected, actual)

		if err := UnscopedDb().Where("photo_name = ? AND photo_quality = ?", "SuccessWith1000", -1).Delete(&entity.Photo{}).Error; err != nil {
			t.Fatal(err)
		}
	})
}
