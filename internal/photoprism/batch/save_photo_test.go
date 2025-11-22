package batch

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func TestSavePhoto(t *testing.T) {
	fixture := entity.PhotoFixtures.Get("Photo01")
	photo := entity.FindPhoto(entity.Photo{PhotoUID: fixture.PhotoUID})
	require.NotNil(t, photo)
	originalTitle := photo.PhotoTitle
	originalFavorite := photo.PhotoFavorite
	originalYear := photo.PhotoYear
	originalMonth := photo.PhotoMonth
	originalDay := photo.PhotoDay
	originalChecked := photo.CheckedAt
	originalEdited := photo.EditedAt

	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := savePhoto(nil)
		require.Error(t, err)
	})
	t.Run("UpdatesCoreFields", func(t *testing.T) {
		values := &PhotosForm{
			PhotoTitle:    String{Value: fmt.Sprintf("Batch %d", time.Now().UnixNano()), Action: ActionUpdate},
			PhotoFavorite: Bool{Value: !photo.PhotoFavorite, Action: ActionUpdate},
			PhotoYear:     Int{Value: 2024, Action: ActionUpdate},
			PhotoMonth:    Int{Value: 12, Action: ActionUpdate},
			PhotoDay:      Int{Value: 31, Action: ActionUpdate},
		}
		frm := &form.Photo{
			PhotoTitle:    values.PhotoTitle.Value,
			PhotoFavorite: values.PhotoFavorite.Value,
			PhotoYear:     values.PhotoYear.Value,
			PhotoMonth:    values.PhotoMonth.Value,
			PhotoDay:      values.PhotoDay.Value,
			TimeZone:      photo.TimeZone,
			TakenAtLocal:  photo.TakenAtLocal,
			TakenSrc:      entity.SrcBatch,
		}

		req, err := NewPhotoSaveRequest(photo, values)
		require.NoError(t, err)
		req.Form = frm

		saved, err := savePhoto(req)
		require.NoError(t, err)
		require.True(t, saved)

		updated := entity.FindPhoto(entity.Photo{PhotoUID: fixture.PhotoUID})
		require.NotNil(t, updated)
		require.Equal(t, values.PhotoTitle.Value, updated.PhotoTitle)
		require.Equal(t, values.PhotoFavorite.Value, updated.PhotoFavorite)
		require.Equal(t, values.PhotoYear.Value, updated.PhotoYear)
		require.Equal(t, values.PhotoMonth.Value, updated.PhotoMonth)
		require.Equal(t, values.PhotoDay.Value, updated.PhotoDay)
		require.Nil(t, updated.CheckedAt)
		require.NotNil(t, updated.EditedAt)

		restorePhoto(t, fixture.PhotoUID, entity.Values{
			"photo_title":    originalTitle,
			"photo_favorite": originalFavorite,
			"photo_year":     originalYear,
			"photo_month":    originalMonth,
			"photo_day":      originalDay,
			"checked_at":     originalChecked,
			"edited_at":      originalEdited,
		})
	})
	t.Run("NoChanges", func(t *testing.T) {
		req, err := NewPhotoSaveRequest(photo, &PhotosForm{})
		require.NoError(t, err)
		saved, err := savePhoto(req)
		require.NoError(t, err)
		require.False(t, saved)
	})
}
