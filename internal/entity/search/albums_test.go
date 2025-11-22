package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func TestAlbumPhotos(t *testing.T) {
	t.Run("SearchWithString", func(t *testing.T) {
		results, err := AlbumPhotos(entity.AlbumFixtures.Get("april-1990"), 2, true)

		if err != nil {
			t.Fatal(err)
		}

		if len(results) < 2 {
			t.Errorf("at least 2 results expected: %d", len(results))
		}
	})
}

func TestUserAlbums(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		query := form.NewAlbumSearch("christmas")
		query.Type = entity.AlbumManual
		result, err := UserAlbums(query, entity.SessionFixtures.Pointer("alice"))

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Christmas 2030", result[0].AlbumTitle)
	})
	t.Run("Visitor", func(t *testing.T) {
		query := form.NewAlbumSearch("christmas")
		query.Type = entity.AlbumFolder
		_, err := UserAlbums(query, entity.SessionFixtures.Pointer("unauthorized"))

		assert.Error(t, err)
		assert.Equal(t, err.Error(), "Permission denied")
	})
	t.Run("GuestAppPassword", func(t *testing.T) {
		query := form.NewAlbumSearch("france")
		query.Type = entity.AlbumMoment
		result, err := UserAlbums(query, entity.SessionFixtures.Pointer("gandalf_app_password_full_access"))

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(result))
	})
}

func TestAlbums(t *testing.T) {
	t.Run("SearchWithString", func(t *testing.T) {
		query := form.NewAlbumSearch("chr")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Christmas 2030", result[0].AlbumTitle)
	})
	t.Run("SearchWithSlug", func(t *testing.T) {
		query := form.NewAlbumSearch("slug:holiday")
		query.Type = entity.AlbumManual
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Holiday 2030", result[0].AlbumTitle)
	})
	t.Run("SearchWithCountry", func(t *testing.T) {
		query := form.NewAlbumSearch("country:ca")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "April 1990", result[0].AlbumTitle)
	})
	t.Run("FavoritesTrue", func(t *testing.T) {
		query := form.NewAlbumSearch("favorite:true")
		query.Count = 100000
		query.Type = entity.AlbumManual

		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Holiday 2030", result[0].AlbumTitle)
	})
	t.Run("EmptyQuery", func(t *testing.T) {
		query := form.NewAlbumSearch("")

		results, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		if len(results) < 3 {
			t.Errorf("at least 3 results expected: %d", len(results))
		}
	})
	t.Run("SearchWithInvalidQueryString", func(t *testing.T) {
		query := form.NewAlbumSearch("xxx:bla")
		result, err := Albums(query)
		assert.Error(t, err, "unknown filter")
		t.Log(result)
	})
	t.Run("SearchWithInvalidQueryString", func(t *testing.T) {
		query := form.NewAlbumSearch("xxx:bla")
		result, err := Albums(query)
		assert.Error(t, err, "unknown filter")
		t.Log(result)
	})
	t.Run("SearchForExistingID", func(t *testing.T) {
		f := form.SearchAlbums{
			Query:    "",
			UID:      "as6sg6bxpogaaba7",
			Slug:     "",
			Title:    "",
			Favorite: false,
			Count:    0,
			Offset:   0,
			Order:    "",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(result))
		assert.Equal(t, "christmas-2030", result[0].AlbumSlug)
	})
	t.Run("SearchWithMultipleFilters", func(t *testing.T) {
		f := form.SearchAlbums{
			Query:    "",
			Type:     "moment",
			Category: "Fun",
			Location: "Favorite Park",
			Title:    "Empty Moment",
			Count:    0,
			Offset:   0,
			Order:    "",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(result))
		assert.Equal(t, "Empty Moment", result[0].AlbumTitle)
	})
	t.Run("SearchForYearMonthDay", func(t *testing.T) {
		f := form.SearchAlbums{
			Year:   "2021",
			Month:  "10",
			Day:    "3",
			Count:  0,
			Offset: 0,
			Order:  "",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(result))
	})
	t.Run("SearchAlbumForYear", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumManual,
			Year:   "2018",
			Month:  "",
			Day:    "",
			Count:  10,
			Offset: 0,
			Order:  "added",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, len(result))
	})
	t.Run("Folders", func(t *testing.T) {
		query := form.NewAlbumSearch("19")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "April 1990", result[0].AlbumTitle)
	})
	t.Run("California", func(t *testing.T) {
		query := form.NewAlbumSearch("california")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("albums: %#v", result)

		assert.GreaterOrEqual(t, 3, len(result))
	})
	t.Run("Blue", func(t *testing.T) {
		query := form.NewAlbumSearch("blue")
		result, err := Albums(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(result))
	})
	t.Run("FolderSortNameReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumFolder,
			Count:   10,
			Offset:  0,
			Order:   "name",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Holiday", result[0].AlbumPath)
		assert.Equal(t, "Holiday", result[0].AlbumTitle)
		assert.Equal(t, "2015/11", result[1].AlbumPath)
		assert.Equal(t, "November 2015", result[1].AlbumTitle)
	})
	t.Run("FolderSortName", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumFolder,
			Count:  10,
			Offset: 0,
			Order:  "name",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "1990/04", result[0].AlbumPath)
		assert.Equal(t, "April 1990", result[0].AlbumTitle)
		assert.Equal(t, "2015/11", result[1].AlbumPath)
		assert.Equal(t, "November 2015", result[1].AlbumTitle)
	})
	t.Run("AlbumSortNameReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumManual,
			Count:   100,
			Offset:  0,
			Order:   "name",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "sale%", result[0].AlbumTitle)
		assert.Equal(t, "Yoga***", result[1].AlbumTitle)
	})
	t.Run("AlbumSortName", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumManual,
			Count:  100,
			Offset: 0,
			Order:  "name",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "%gold", result[0].AlbumTitle)
		assert.Equal(t, "'Family", result[1].AlbumTitle)
	})
	t.Run("SortByCount", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumManual,
			Count:  100,
			Offset: 0,
			Order:  "count",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[0].PhotoCount, result[1].PhotoCount)
	})
	t.Run("SortByCountReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumManual,
			Count:   100,
			Offset:  0,
			Order:   "count",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[1].PhotoCount, result[0].PhotoCount)
	})
	t.Run("AlbumSortByNewest", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumManual,
			Count:  100,
			Offset: 0,
			Order:  "newest",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[0].AlbumUID, result[1].AlbumUID)
	})
	t.Run("AlbumSortByNewestReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumManual,
			Count:   100,
			Offset:  0,
			Order:   "newest",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[1].AlbumUID, result[0].AlbumUID)
	})
	t.Run("MomentSortByNewest", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumMoment,
			Count:  100,
			Offset: 2,
			Order:  "newest",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[0].AlbumYear, result[1].AlbumYear)
	})
	t.Run("MomentSortByNewestReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumMoment,
			Count:   100,
			Offset:  2,
			Order:   "newest",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[1].AlbumYear, result[0].AlbumYear)
	})
	t.Run("FolderSortByNewest", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumFolder,
			Count:  100,
			Offset: 0,
			Order:  "newest",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[0].AlbumYear, result[1].AlbumYear)
	})
	t.Run("FolderSortByNewestReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumFolder,
			Count:   100,
			Offset:  0,
			Order:   "newest",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[1].AlbumYear, result[0].AlbumYear)
	})
	t.Run("AlbumSortByOldest", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumManual,
			Count:  100,
			Offset: 0,
			Order:  "oldest",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[1].AlbumUID, result[0].AlbumUID)
	})
	t.Run("AlbumSortByOldestReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumManual,
			Count:   100,
			Offset:  0,
			Order:   "oldest",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[0].AlbumUID, result[1].AlbumUID)
	})
	t.Run("MomentSortByOldest", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumMoment,
			Count:  100,
			Offset: 2,
			Order:  "oldest",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[1].AlbumYear, result[0].AlbumYear)
	})
	t.Run("MomentSortByOldestReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumMoment,
			Count:   100,
			Offset:  2,
			Order:   "oldest",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[0].AlbumYear, result[1].AlbumYear)
	})
	t.Run("FolderSortByOldest", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumFolder,
			Count:  100,
			Offset: 0,
			Order:  "oldest",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[1].AlbumYear, result[0].AlbumYear)
	})
	t.Run("FolderSortByOldestReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumFolder,
			Count:   100,
			Offset:  0,
			Order:   "oldest",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[0].AlbumYear, result[1].AlbumYear)
	})
	t.Run("MomentSortByEdited", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumMoment,
			Count:  100,
			Offset: 0,
			Order:  "edited",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[0].UpdatedAt, result[1].UpdatedAt)
	})
	t.Run("MomentSortByEditedReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumMoment,
			Count:   100,
			Offset:  0,
			Order:   "edited",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, result[0].UpdatedAt, result[1].UpdatedAt)
	})
	t.Run("MomentSortByPlace", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumMoment,
			Count:  100,
			Offset: 0,
			Order:  "place",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Favorite Park", result[0].AlbumLocation)
		assert.Equal(t, "France", result[1].AlbumLocation)
	})
	t.Run("FolderSortByPath", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumFolder,
			Count:  100,
			Offset: 0,
			Order:  "path",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "1990/04", result[0].AlbumPath)
		assert.Equal(t, "2015/11", result[1].AlbumPath)
	})
	t.Run("FolderSortBySlug", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumFolder,
			Count:  100,
			Offset: 0,
			Order:  "slug",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "april-1990", result[0].AlbumSlug)
		assert.Equal(t, "holiday", result[1].AlbumSlug)
	})
	t.Run("FolderSortBySlugReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumFolder,
			Count:   100,
			Offset:  0,
			Order:   "slug",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "november-2015", result[0].AlbumSlug)
		assert.Equal(t, "holiday", result[1].AlbumSlug)
	})
	t.Run("FolderSortByFavorites", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumFolder,
			Count:  100,
			Offset: 0,
			Order:  "favorites",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, true, result[0].AlbumFavorite)
	})
	t.Run("FolderSortByFavoritesReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumFolder,
			Count:   100,
			Offset:  0,
			Order:   "favorites",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, false, result[0].AlbumFavorite)
	})
	t.Run("MomentSortByFavorites", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumMoment,
			Count:  100,
			Offset: 0,
			Order:  "favorites",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, true, result[0].AlbumFavorite)
	})
	t.Run("MomentSortByFavoritesReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumMoment,
			Count:   100,
			Offset:  0,
			Order:   "favorites",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, false, result[0].AlbumFavorite)
	})
	t.Run("AlbumSortByFavorites", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:   entity.AlbumManual,
			Count:  100,
			Offset: 0,
			Order:  "favorites",
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, true, result[0].AlbumFavorite)
	})
	t.Run("AlbumSortByFavoritesReverse", func(t *testing.T) {
		f := form.SearchAlbums{
			Type:    entity.AlbumManual,
			Count:   100,
			Offset:  0,
			Order:   "favorites",
			Reverse: true,
		}

		result, err := Albums(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, false, result[0].AlbumFavorite)
	})
}
