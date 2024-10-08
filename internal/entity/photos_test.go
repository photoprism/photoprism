package entity

import (
	"sort"
	"testing"

	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestPhotos_Photos(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {

		photo1 := PhotoFixtures.Get("Photo08")
		photo2 := PhotoFixtures.Get("Photo07")

		photos := Photos{photo1, photo2}

		r := photos.Photos()

		assert.Equal(t, 2, len(r))
	})
}

// By is the type of a "less" function that defines the ordering of its Photo arguments.
type By func(p1, p2 *Photo) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(photos []Photo) {
	ps := &photoSorter{
		photos: photos,
		by:     by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ps)
}

// photoSorter joins a By function and a slice of Photos to be sorted.
type photoSorter struct {
	photos []Photo
	by     func(p1, p2 *Photo) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *photoSorter) Len() int {
	return len(s.photos)
}

// Swap is part of sort.Interface.
func (s *photoSorter) Swap(i, j int) {
	s.photos[i], s.photos[j] = s.photos[j], s.photos[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *photoSorter) Less(i, j int) bool {
	return s.by(&s.photos[i], &s.photos[j])
}

func TestPhotos_UnscopedSearch(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		photos := Photos{}

		if res := UnscopedSearchPhotos(&photos, "photo_uid in (?, ?, ?)", PhotoFixtures.Get("Photo07").PhotoUID, PhotoFixtures.Get("Photo08").PhotoUID, PhotoFixtures.Get("Photo18").PhotoUID); res.Error != nil {
			assert.Nil(t, res.Error)
			t.FailNow()
		}
		photo1 := PhotoFixtures.Get("Photo08")
		photo2 := PhotoFixtures.Get("Photo07")
		photo3 := PhotoFixtures.Get("Photo18")

		expectedPhotos := Photos{photo2, photo1, photo3}

		r := photos.Photos()

		assert.Equal(t, 3, len(r))

		uuid := func(p1, p2 *Photo) bool {
			return p1.PhotoUID < p2.PhotoUID
		}

		// Make sure we don't fail because of sorting
		By(uuid).Sort(photos)
		By(uuid).Sort(expectedPhotos)

		// Only check items that are preloaded
		// Except Labels as they are filtered.
		for i := 0; i < 3; i++ {
			assert.Equal(t, expectedPhotos[i].ID, photos[i].ID)
			assert.Equal(t, expectedPhotos[i].UUID, photos[i].UUID)
			assert.Equal(t, expectedPhotos[i].TakenAt, photos[i].TakenAt)
			assert.Equal(t, expectedPhotos[i].TakenSrc, photos[i].TakenSrc)
			assert.Equal(t, expectedPhotos[i].PhotoUID, photos[i].PhotoUID)
			assert.Equal(t, expectedPhotos[i].PhotoPath, photos[i].PhotoPath)
			assert.Equal(t, expectedPhotos[i].Camera, photos[i].Camera)
			assert.Equal(t, expectedPhotos[i].CameraID, photos[i].CameraID)
			assert.Equal(t, expectedPhotos[i].Lens, photos[i].Lens)
			assert.Equal(t, expectedPhotos[i].LensID, photos[i].LensID)
			assert.Equal(t, expectedPhotos[i].Place.PlaceLabel, photos[i].Place.PlaceLabel)
			assert.Equal(t, expectedPhotos[i].PlaceID, photos[i].PlaceID)
			assert.Equal(t, expectedPhotos[i].Cell.CellName, photos[i].Cell.CellName) // CellName as PhotoCount can cause this to fail
			assert.Equal(t, expectedPhotos[i].CellID, photos[i].CellID)
		}
	})

	t.Run("Nothing Found", func(t *testing.T) {
		photos := Photos{}

		res := &gorm.DB{}
		if res = UnscopedSearchPhotos(&photos, "photo_uid in (?, ?, ?)", rnd.UUID(), rnd.UUID(), rnd.UUID()); res.Error != nil {
			assert.Nil(t, res.Error)
			t.FailNow()
		}

		assert.Equal(t, int64(0), res.RowsAffected)
	})

	t.Run("Error", func(t *testing.T) {
		photos := Photos{}

		res := &gorm.DB{}
		if res = UnscopedSearchPhotos(&photos, "photo_uids in (?, ?, ?)", rnd.UUID(), rnd.UUID(), rnd.UUID()); res.Error == nil {
			assert.NotNil(t, res.Error)
			t.FailNow()
		}
		assert.Error(t, res.Error)
		assert.ErrorContains(t, res.Error, "photo_uids")
		assert.Equal(t, int64(0), res.RowsAffected)
	})
}

func TestPhotos_ScopedSearch(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		photos := Photos{}

		if res := ScopedSearchPhotos(&photos, "photo_uid in (?, ?, ?)", PhotoFixtures.Get("Photo07").PhotoUID, PhotoFixtures.Get("Photo08").PhotoUID, PhotoFixtures.Get("Photo18").PhotoUID); res.Error != nil {
			assert.Nil(t, res.Error)
			t.FailNow()
		}
		photo1 := PhotoFixtures.Get("Photo08")
		photo2 := PhotoFixtures.Get("Photo07")

		expectedPhotos := Photos{photo2, photo1}

		r := photos.Photos()

		assert.Equal(t, 2, len(r))

		uuid := func(p1, p2 *Photo) bool {
			return p1.PhotoUID < p2.PhotoUID
		}

		// Make sure we don't fail because of sorting
		By(uuid).Sort(photos)
		By(uuid).Sort(expectedPhotos)

		// Only check items that are preloaded
		// Except Labels as they are filtered.
		for i := 0; i < 2; i++ {
			assert.Equal(t, expectedPhotos[i].ID, photos[i].ID)
			assert.Equal(t, expectedPhotos[i].UUID, photos[i].UUID)
			assert.Equal(t, expectedPhotos[i].TakenAt, photos[i].TakenAt)
			assert.Equal(t, expectedPhotos[i].TakenSrc, photos[i].TakenSrc)
			assert.Equal(t, expectedPhotos[i].PhotoUID, photos[i].PhotoUID)
			assert.Equal(t, expectedPhotos[i].PhotoPath, photos[i].PhotoPath)
			assert.Equal(t, expectedPhotos[i].Camera, photos[i].Camera)
			assert.Equal(t, expectedPhotos[i].CameraID, photos[i].CameraID)
			assert.Equal(t, expectedPhotos[i].Lens, photos[i].Lens)
			assert.Equal(t, expectedPhotos[i].LensID, photos[i].LensID)
			assert.Equal(t, expectedPhotos[i].Place.PlaceLabel, photos[i].Place.PlaceLabel)
			assert.Equal(t, expectedPhotos[i].PlaceID, photos[i].PlaceID)
			assert.Equal(t, expectedPhotos[i].Cell.CellName, photos[i].Cell.CellName) // CellName as PhotoCount can cause this to fail
			assert.Equal(t, expectedPhotos[i].CellID, photos[i].CellID)
		}
	})

	t.Run("Nothing Found", func(t *testing.T) {
		photos := Photos{}

		res := &gorm.DB{}
		if res = ScopedSearchPhotos(&photos, "photo_uid in (?, ?, ?)", rnd.UUID(), rnd.UUID(), rnd.UUID()); res.Error != nil {
			assert.Nil(t, res.Error)
			t.FailNow()
		}

		assert.Equal(t, int64(0), res.RowsAffected)
	})

	t.Run("Error", func(t *testing.T) {
		photos := Photos{}

		res := &gorm.DB{}
		if res = ScopedSearchPhotos(&photos, "photo_uids in (?, ?, ?)", rnd.UUID(), rnd.UUID(), rnd.UUID()); res.Error == nil {
			assert.NotNil(t, res.Error)
			t.FailNow()
		}
		assert.Error(t, res.Error)
		assert.ErrorContains(t, res.Error, "photo_uids")
		assert.Equal(t, int64(0), res.RowsAffected)
	})
}
