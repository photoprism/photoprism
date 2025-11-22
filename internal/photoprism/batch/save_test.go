package batch

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
)

// TestSavePhotos covers SavePhotos scenarios.
func TestSavePhotos(t *testing.T) {
	t.Run("UpdatesTitleAndFavorite", func(t *testing.T) {
		fixture := entity.PhotoFixtures.Get("Photo01")
		photo := entity.FindPhoto(entity.Photo{PhotoUID: fixture.PhotoUID})
		require.NotNil(t, photo)

		originalTitle := photo.PhotoTitle
		originalTitleSrc := photo.TitleSrc
		originalFavorite := photo.PhotoFavorite
		originalChecked := photo.CheckedAt
		originalEdited := photo.EditedAt

		values := &PhotosForm{
			PhotoTitle:    String{Value: originalTitle + " (Batch)", Action: ActionUpdate},
			PhotoFavorite: Bool{Value: !originalFavorite, Action: ActionUpdate},
		}

		req, err := NewPhotoSaveRequest(photo, values)
		require.NoError(t, err)

		results, err := SavePhotos([]*PhotoSaveRequest{req})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.True(t, results[0])

		updated := entity.FindPhoto(entity.Photo{PhotoUID: fixture.PhotoUID})
		require.NotNil(t, updated)
		assert.Equal(t, values.PhotoTitle.Value, updated.PhotoTitle)
		assert.Equal(t, entity.SrcBatch, updated.TitleSrc)
		assert.Equal(t, values.PhotoFavorite.Value, updated.PhotoFavorite)
		assert.Nil(t, updated.CheckedAt)
		assert.NotNil(t, updated.EditedAt)

		restorePhoto(t, fixture.PhotoUID, entity.Values{
			"photo_title":    originalTitle,
			"title_src":      originalTitleSrc,
			"photo_favorite": originalFavorite,
			"checked_at":     originalChecked,
			"edited_at":      originalEdited,
		})
	})

	t.Run("UpdatesDateFields", func(t *testing.T) {
		fixture := entity.PhotoFixtures.Get("Photo02")
		photo := entity.FindPhoto(entity.Photo{PhotoUID: fixture.PhotoUID})
		require.NotNil(t, photo)

		originalYear := photo.PhotoYear
		originalMonth := photo.PhotoMonth
		originalDay := photo.PhotoDay
		originalTakenAt := photo.TakenAt
		originalTakenAtLocal := photo.TakenAtLocal
		originalTimeZone := photo.TimeZone
		originalTakenSrc := photo.TakenSrc
		originalChecked := photo.CheckedAt
		originalEdited := photo.EditedAt

		values := &PhotosForm{
			PhotoYear:  Int{Value: 2020, Action: ActionUpdate},
			PhotoMonth: Int{Value: 5, Action: ActionUpdate},
			PhotoDay:   Int{Value: 15, Action: ActionUpdate},
		}

		req, err := NewPhotoSaveRequest(photo, values)
		require.NoError(t, err)

		results, err := SavePhotos([]*PhotoSaveRequest{req})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.True(t, results[0])

		updated := entity.FindPhoto(entity.Photo{PhotoUID: fixture.PhotoUID})
		require.NotNil(t, updated)
		assert.Equal(t, 2020, updated.PhotoYear)
		assert.Equal(t, 5, updated.PhotoMonth)
		assert.Equal(t, 15, updated.PhotoDay)
		assert.Equal(t, entity.SrcBatch, updated.TakenSrc)
		assert.Nil(t, updated.CheckedAt)

		restorePhoto(t, fixture.PhotoUID, entity.Values{
			"photo_year":     originalYear,
			"photo_month":    originalMonth,
			"photo_day":      originalDay,
			"taken_at":       originalTakenAt,
			"taken_at_local": originalTakenAtLocal,
			"time_zone":      originalTimeZone,
			"taken_src":      originalTakenSrc,
			"checked_at":     originalChecked,
			"edited_at":      originalEdited,
		})
	})

	t.Run("RemovesStrings", func(t *testing.T) {
		fixture := entity.PhotoFixtures.Get("Photo03")
		photo := entity.FindPhoto(entity.Photo{PhotoUID: fixture.PhotoUID})
		require.NotNil(t, photo)

		original := entity.Values{
			"photo_title":           photo.PhotoTitle,
			"title_src":             photo.TitleSrc,
			"photo_caption":         photo.PhotoCaption,
			"caption_src":           photo.CaptionSrc,
			"details_subject":       photo.GetDetails().Subject,
			"details_subject_src":   photo.GetDetails().SubjectSrc,
			"details_artist":        photo.GetDetails().Artist,
			"details_artist_src":    photo.GetDetails().ArtistSrc,
			"details_copyright":     photo.GetDetails().Copyright,
			"details_copyright_src": photo.GetDetails().CopyrightSrc,
			"details_license":       photo.GetDetails().License,
			"details_license_src":   photo.GetDetails().LicenseSrc,
			"checked_at":            photo.CheckedAt,
			"edited_at":             photo.EditedAt,
		}

		setValues := &PhotosForm{
			PhotoTitle:       String{Value: "Batch Title", Action: ActionUpdate},
			PhotoCaption:     String{Value: "Batch Caption", Action: ActionUpdate},
			DetailsSubject:   String{Value: "Batch Subject", Action: ActionUpdate},
			DetailsArtist:    String{Value: "Batch Artist", Action: ActionUpdate},
			DetailsCopyright: String{Value: "Batch Copyright", Action: ActionUpdate},
			DetailsLicense:   String{Value: "Batch License", Action: ActionUpdate},
		}
		setReq, err := NewPhotoSaveRequest(photo, setValues)
		require.NoError(t, err)
		_, err = SavePhotos([]*PhotoSaveRequest{setReq})
		require.NoError(t, err)

		removeValues := &PhotosForm{
			PhotoTitle:       String{Action: ActionRemove},
			PhotoCaption:     String{Action: ActionRemove},
			DetailsSubject:   String{Action: ActionRemove},
			DetailsArtist:    String{Action: ActionRemove},
			DetailsCopyright: String{Action: ActionRemove},
			DetailsLicense:   String{Action: ActionRemove},
			PhotoYear:        Int{Value: -1, Action: ActionUpdate},
			PhotoMonth:       Int{Value: -1, Action: ActionUpdate},
			PhotoDay:         Int{Value: -1, Action: ActionUpdate},
			PhotoAltitude:    Int{Value: 0, Action: ActionUpdate},
		}
		removeReq, err := NewPhotoSaveRequest(photo, removeValues)
		require.NoError(t, err)
		results, err := SavePhotos([]*PhotoSaveRequest{removeReq})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.True(t, results[0])

		updated := entity.FindPhoto(entity.Photo{PhotoUID: fixture.PhotoUID})
		require.NotNil(t, updated)
		assert.Equal(t, "", updated.PhotoTitle)
		assert.Equal(t, "", updated.PhotoCaption)
		assert.Equal(t, "", updated.GetDetails().Subject)
		assert.Equal(t, "", updated.GetDetails().Artist)
		assert.Equal(t, "", updated.GetDetails().Copyright)
		assert.Equal(t, "", updated.GetDetails().License)

		restorePhoto(t, fixture.PhotoUID, original)
	})
}

// TestNewPhotoSaveRequest ensures the helper validates inputs before building requests.
func TestNewPhotoSaveRequest(t *testing.T) {
	t.Run("NilValues", func(t *testing.T) {
		fixture := entity.PhotoFixtures.Get("Photo01")
		photo := entity.FindPhoto(entity.Photo{PhotoUID: fixture.PhotoUID})
		require.NotNil(t, photo)

		req, err := NewPhotoSaveRequest(photo, nil)
		assert.Nil(t, req)
		assert.Error(t, err)
	})

	t.Run("BuildsRequest", func(t *testing.T) {
		fixture := entity.PhotoFixtures.Get("Photo02")
		photo := entity.FindPhoto(entity.Photo{PhotoUID: fixture.PhotoUID})
		require.NotNil(t, photo)

		values := &PhotosForm{
			PhotoTitle: String{Value: "Helper", Action: ActionUpdate},
		}

		req, err := NewPhotoSaveRequest(photo, values)
		require.NoError(t, err)
		require.NotNil(t, req)
		assert.Equal(t, photo, req.Photo)
		assert.Equal(t, values, req.Values)
		assert.NotNil(t, req.Form)
		assert.Equal(t, "Helper", req.Form.PhotoTitle)
	})
}

// TestPreparePhotoSaveRequests verifies helper behavior.
func TestPreparePhotoSaveRequests(t *testing.T) {
	t.Run("NilValues", func(t *testing.T) {
		preloaded := map[string]*entity.Photo{}
		requests, updated, stats := PreparePhotoSaveRequests(nil, preloaded, nil)
		assert.Nil(t, requests)
		assert.Equal(t, preloaded, updated)
		assert.Equal(t, MutationStats{}, stats)
	})

	t.Run("LoadsMissingPhoto", func(t *testing.T) {
		fixture := entity.PhotoFixtures.Get("Photo01")
		values := &PhotosForm{PhotoTitle: String{Value: "Prepared", Action: ActionUpdate}}
		photos := search.PhotoResults{{PhotoUID: fixture.PhotoUID}}

		requests, updated, stats := PreparePhotoSaveRequests(photos, nil, values)
		require.Len(t, requests, 1)
		assert.NotNil(t, requests[0].Photo)
		assert.Equal(t, "Prepared", requests[0].Form.PhotoTitle)
		assert.Contains(t, updated, fixture.PhotoUID)
		assert.Equal(t, MutationStats{}, stats)
	})

	t.Run("SkipsMissing", func(t *testing.T) {
		values := &PhotosForm{PhotoTitle: String{Value: "Prepared", Action: ActionUpdate}}
		photos := search.PhotoResults{{PhotoUID: "pt_does_not_exist"}}

		requests, updated, stats := PreparePhotoSaveRequests(photos, nil, values)
		assert.Len(t, requests, 0)
		assert.Empty(t, updated)
		assert.Equal(t, MutationStats{}, stats)
	})
}

// TestPrepareAndSavePhotos verifies the full helper workflow.
func TestPrepareAndSavePhotos(t *testing.T) {
	t.Run("NilValues", func(t *testing.T) {
		result, err := PrepareAndSavePhotos(nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.NotNil(t, result.Preloaded)
		assert.Len(t, result.Requests, 0)
		assert.Len(t, result.Results, 0)
	})

	t.Run("PersistsChanges", func(t *testing.T) {
		fixture := entity.PhotoFixtures.Get("Photo02")
		photo := entity.FindPhoto(entity.Photo{PhotoUID: fixture.PhotoUID})
		require.NotNil(t, photo)

		originalFavorite := photo.PhotoFavorite

		values := &PhotosForm{
			PhotoFavorite: Bool{Value: !originalFavorite, Action: ActionUpdate},
		}
		photos := search.PhotoResults{{PhotoUID: fixture.PhotoUID}}

		result, err := PrepareAndSavePhotos(photos, nil, values)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Requests, 1)
		require.Len(t, result.Results, 1)
		assert.True(t, result.Results[0])
		assert.True(t, result.SavedAny)
		assert.Equal(t, 1, result.UpdatedCount)
		assert.Contains(t, result.Preloaded, fixture.PhotoUID)

		updated := entity.FindPhoto(entity.Photo{PhotoUID: fixture.PhotoUID})
		require.NotNil(t, updated)
		assert.Equal(t, !originalFavorite, updated.PhotoFavorite)

		restorePhoto(t, fixture.PhotoUID, entity.Values{
			"photo_favorite": originalFavorite,
		})
	})
}

// restorePhoto rewinds DB state for the provided fixture so tests stay isolated.
func restorePhoto(t *testing.T, photoUID string, values entity.Values) {
	t.Helper()
	if values == nil {
		return
	}
	detailUpdates := entity.Values{}
	detailKeys := []string{
		"details_subject", "details_subject_src",
		"details_artist", "details_artist_src",
		"details_copyright", "details_copyright_src",
		"details_license", "details_license_src",
	}
	for _, k := range detailKeys {
		if v, ok := values[k]; ok {
			detailUpdates[strings.TrimPrefix(k, "details_")] = v
			delete(values, k)
		}
	}
	if err := entity.Db().Model(&entity.Photo{}).Where("photo_uid = ?", photoUID).Updates(values).Error; err != nil {
		t.Fatalf("failed to restore photo %s: %v", photoUID, err)
	}
	if len(detailUpdates) > 0 {
		if photo := entity.FindPhoto(entity.Photo{PhotoUID: photoUID}); photo != nil {
			if err := entity.Db().Model(photo.GetDetails()).Updates(detailUpdates).Error; err != nil {
				t.Fatalf("failed to restore photo details %s: %v", photoUID, err)
			}
		}
	}
}
