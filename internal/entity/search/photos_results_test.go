package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_Ids(t *testing.T) {
	r := Photo{
		ID:           1111198,
		CreatedAt:    time.Time{},
		UpdatedAt:    time.Time{},
		DeletedAt:    &time.Time{},
		TakenAt:      time.Time{},
		TakenAtLocal: time.Time{},
		PhotoUID:     "ps6sg6be2lvl0o98",
	}

	assert.Equal(t, uint(1111198), r.GetID())
	assert.True(t, r.HasID())
	assert.Equal(t, "ps6sg6be2lvl0o98", r.GetUID())
}

func TestPhoto_Approve(t *testing.T) {
	t.Run("EmptyPhoto", func(t *testing.T) {
		r := Photo{}
		err := r.Approve()

		assert.Error(t, err)
	})
	t.Run("PhotoNotInReview", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0r41",
			PhotoQuality: 4,
		}

		err := r.Approve()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 4, r.PhotoQuality)
	})
	t.Run("Approve", func(t *testing.T) {
		r := Photo{
			ID:           100028476,
			CreatedAt:    time.Time{},
			UpdatedAt:    time.Time{},
			DeletedAt:    &time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0j76",
			PhotoQuality: 2,
		}

		err := r.Approve()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, r.PhotoQuality)
		assert.Nil(t, r.DeletedAt)
		assert.NotNil(t, r.EditedAt)
	})
}

func TestPhoto_Restore(t *testing.T) {
	t.Run("EmptyPhoto", func(t *testing.T) {
		r := Photo{}

		err := r.Restore()

		assert.Error(t, err)
	})
	t.Run("PhotoNotInArchive", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0r41",
		}

		err := r.Restore()

		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, r.DeletedAt)
	})
	t.Run("Restore", func(t *testing.T) {
		r := Photo{
			ID:           100028476,
			CreatedAt:    time.Time{},
			UpdatedAt:    time.Time{},
			DeletedAt:    &time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0j76",
			PhotoQuality: 2,
		}

		assert.NotNil(t, r.DeletedAt)

		err := r.Restore()

		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, r.DeletedAt)
	})
}

func TestPhoto_IsPlayable(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0r41",
			PhotoType:    "live",
		}

		assert.True(t, r.IsPlayable())
	})
	t.Run("False", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0r41",
			PhotoType:    "image",
		}

		assert.False(t, r.IsPlayable())
	})
}

func TestPhotoResults_Photos(t *testing.T) {
	photo1 := Photo{
		ID:           1111154,
		CreatedAt:    time.Time{},
		TakenAt:      time.Time{},
		TakenAtLocal: time.Time{},
		TakenSrc:     "",
		TimeZone:     "",
		PhotoUID:     "ps6sg6be2lvl0r41",
		PhotoType:    "live",
	}

	photo2 := Photo{
		ID:           1111155,
		CreatedAt:    time.Time{},
		TakenAt:      time.Time{},
		TakenAtLocal: time.Time{},
		TakenSrc:     "",
		TimeZone:     "",
		PhotoUID:     "ps6sg6be2lvl0986",
		PhotoType:    "image",
	}

	r := PhotoResults{photo1, photo2}

	assert.Equal(t, 2, len(r.Photos()))
}

func TestPhotosResults_Merged(t *testing.T) {
	result1 := Photo{
		ID:               111111,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
		DeletedAt:        &time.Time{},
		TakenAt:          time.Time{},
		TakenAtLocal:     time.Time{},
		TakenSrc:         "",
		TimeZone:         "",
		PhotoUID:         "",
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
		CameraType:       "",
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
		PhotoUID:         "",
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
		CameraType:       "",
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

	merged, count, err := results.Merge()

	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, count)
	t.Log(merged)
}
func TestPhotosResults_UIDs(t *testing.T) {
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
		CameraType:       "",
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
		CameraType:       "",
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

	result := results.UIDs()
	assert.Equal(t, []string{"123", "456"}, result)
}

func TestPhotosResult_ShareFileName(t *testing.T) {
	t.Run("WithTitle", func(t *testing.T) {
		result1 := Photo{
			ID:               111111,
			CreatedAt:        time.Time{},
			UpdatedAt:        time.Time{},
			DeletedAt:        &time.Time{},
			TakenAt:          time.Date(2015, 11, 11, 9, 7, 18, 0, time.UTC),
			TakenAtLocal:     time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC),
			TakenSrc:         "",
			TimeZone:         "",
			PhotoUID:         "uid123",
			PhotoPath:        "",
			PhotoName:        "",
			PhotoTitle:       "PhotoTitle123",
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
			CameraType:       "",
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

		r := result1.ShareBase(0)
		assert.Contains(t, r, "20131111-090718-Phototitle123")
	})
	t.Run("NoTitle", func(t *testing.T) {
		result1 := Photo{
			ID:               111111,
			CreatedAt:        time.Time{},
			UpdatedAt:        time.Time{},
			DeletedAt:        &time.Time{},
			TakenAt:          time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC),
			TakenAtLocal:     time.Date(2015, 11, 11, 9, 7, 18, 0, time.UTC),
			TakenSrc:         "",
			TimeZone:         "",
			PhotoUID:         "uid123",
			PhotoPath:        "",
			PhotoName:        "",
			PhotoTitle:       "",
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
			CameraType:       "",
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

		r := result1.ShareBase(0)
		assert.Contains(t, r, "20151111-090718-uid123")
	})

	t.Run("SeqGreater0", func(t *testing.T) {
		result1 := Photo{
			ID:               111111,
			CreatedAt:        time.Time{},
			UpdatedAt:        time.Time{},
			DeletedAt:        &time.Time{},
			TakenAt:          time.Date(2022, 11, 11, 9, 7, 18, 0, time.UTC),
			TakenAtLocal:     time.Date(2022, 11, 11, 9, 7, 18, 0, time.UTC),
			TakenSrc:         "",
			TimeZone:         "",
			PhotoUID:         "uid123",
			PhotoPath:        "",
			PhotoName:        "",
			PhotoTitle:       "PhotoTitle123",
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
			CameraType:       "",
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

		r := result1.ShareBase(3)
		assert.Contains(t, r, "20221111-090718-Phototitle123 (3)")
	})
}
