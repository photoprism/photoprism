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

func TestSavePhoto_ControlCharacters(t *testing.T) {
	m := entity.NewPhoto(false)
	require.NoError(t, m.Create())
	t.Cleanup(func() { _, _ = m.DeletePermanently() })

	photo := entity.FindPhoto(m)
	require.NotNil(t, photo)

	values := &PhotosForm{
		PhotoTitle:   String{Value: "Batch\x00 Title", Action: ActionUpdate},
		PhotoCountry: String{Value: "d\x00e", Action: ActionUpdate},
	}

	req, err := NewPhotoSaveRequest(photo, values)
	require.NoError(t, err)

	saved, err := savePhoto(req)
	require.NoError(t, err)
	require.True(t, saved)

	updated := entity.FindPhoto(m)
	require.NotNil(t, updated)
	require.Equal(t, "Batch Title", updated.PhotoTitle)
	require.Equal(t, "de", updated.PhotoCountry)
}

// TestSavePhoto_TimeZone checks that batch date and time zone changes keep each photo's local time.
func TestSavePhoto_TimeZone(t *testing.T) {
	cases := []struct {
		name, zone, taken, local string
		day                      int
		values                   PhotosForm
		expectedTaken            string
		expectedLocal            string
		expectedDate             []int
	}{
		{"FromUTCKeepsWallTime", "UTC", "2025-06-30T23:30:00Z", "2025-06-30T23:30:00Z", 30, PhotosForm{TimeZone: String{Value: "Europe/Berlin", Action: ActionUpdate}}, "2025-06-30T21:30:00Z", "2025-06-30T23:30:00Z", []int{2025, 6, 30}},
		{"FromUTCUnknownDay", "UTC", "2020-01-01T03:00:00Z", "2020-01-01T03:00:00Z", -1, PhotosForm{TimeZone: String{Value: "America/New_York", Action: ActionUpdate}}, "2020-01-01T08:00:00Z", "2020-01-01T03:00:00Z", []int{2020, 1, -1}},
		{"FromUTCWithDateChange", "UTC", "2025-06-30T23:30:00Z", "2025-06-30T23:30:00Z", 30, PhotosForm{TimeZone: String{Value: "Europe/Berlin", Action: ActionUpdate}, PhotoDay: Int{Value: 2, Action: ActionUpdate}}, "2025-06-02T21:30:00Z", "2025-06-02T23:30:00Z", []int{2025, 6, 2}},
		{"FromNamedKeepsWallTime", "Europe/Berlin", "2025-06-30T12:00:00Z", "2025-06-30T14:00:00Z", 30, PhotosForm{TimeZone: String{Value: "America/New_York", Action: ActionUpdate}}, "2025-06-30T18:00:00Z", "2025-06-30T14:00:00Z", []int{2025, 6, 30}},
		{"ToUTCKeepsWallTime", "Europe/Berlin", "2025-06-30T12:00:00Z", "2025-06-30T14:00:00Z", 30, PhotosForm{TimeZone: String{Value: "UTC", Action: ActionUpdate}}, "2025-06-30T14:00:00Z", "2025-06-30T14:00:00Z", []int{2025, 6, 30}},
		{"MixedZoneDateInUTC", "UTC", "2025-06-30T23:30:00Z", "2025-06-30T23:30:00Z", 30, PhotosForm{PhotoDay: Int{Value: 2, Action: ActionUpdate}}, "2025-06-02T23:30:00Z", "2025-06-02T23:30:00Z", []int{2025, 6, 2}},
		{"MixedZoneDateInNamedZone", "Europe/Berlin", "2025-06-30T12:00:00Z", "2025-06-30T14:00:00Z", 30, PhotosForm{PhotoDay: Int{Value: 2, Action: ActionUpdate}}, "2025-06-02T12:00:00Z", "2025-06-02T14:00:00Z", []int{2025, 6, 2}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := entity.NewPhoto(false)
			m.TakenAt, _ = time.Parse(time.RFC3339, tc.taken)
			m.TakenAtLocal, _ = time.Parse(time.RFC3339, tc.local)
			m.TimeZone, m.TakenSrc = tc.zone, entity.SrcManual
			m.PhotoYear, m.PhotoMonth, m.PhotoDay = m.TakenAtLocal.Year(), int(m.TakenAtLocal.Month()), tc.day
			require.NoError(t, m.Create())
			t.Cleanup(func() { _, _ = m.DeletePermanently() })

			photo := entity.FindPhoto(m)
			require.NotNil(t, photo)
			req, err := NewPhotoSaveRequest(photo, &tc.values)
			require.NoError(t, err)
			saved, err := savePhoto(req)
			require.NoError(t, err)
			require.True(t, saved)

			updated := entity.FindPhoto(m)
			require.NotNil(t, updated)
			require.Equal(t, tc.expectedTaken, updated.TakenAt.Format(time.RFC3339))
			require.Equal(t, tc.expectedLocal, updated.TakenAtLocal.Format(time.RFC3339))
			require.Equal(t, tc.expectedDate, []int{updated.PhotoYear, updated.PhotoMonth, updated.PhotoDay})
		})
	}
}
