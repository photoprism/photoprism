package query

import (
	"testing"

	"github.com/dustin/go-humanize/english"
	"github.com/photoprism/photoprism/internal/entity"

	"github.com/stretchr/testify/assert"
)

func TestMomentsTime(t *testing.T) {
	t.Run("PublicOnly", func(t *testing.T) {
		results, err := MomentsTime(1, true)

		if err != nil {
			t.Fatal(err)
		}
		if len(results) < 4 {
			t.Error("at least 4 results expected")
		}

		t.Logf("MomentsTime %+v", results)

		for _, moment := range results {
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
			t.Logf("Title Slug: %s", moment.TitleSlug())

			assert.Len(t, moment.Country, 0)
			assert.GreaterOrEqual(t, moment.Year, 1990)
			assert.LessOrEqual(t, moment.Year, 2800)
			assert.GreaterOrEqual(t, moment.Month, 1)
			assert.LessOrEqual(t, moment.Month, 12)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			assert.Regexp(t, "[a-zA-Z]+ [0-9]+", moment.Title())
			assert.Regexp(t, "[a-z]+\\-[0-9]+", moment.Slug())
			assert.Regexp(t, "[a-z]+\\-[0-9]+", moment.TitleSlug())
		}
	})
	t.Run("IncludePrivate", func(t *testing.T) {
		results, err := MomentsTime(1, false)

		if err != nil {
			t.Fatal(err)
		}
		if len(results) < 4 {
			t.Error("at least 4 results expected")
		}

		t.Logf("MomentsTime %+v", results)

		for _, moment := range results {
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
			t.Logf("Title Slug: %s", moment.TitleSlug())

			assert.Len(t, moment.Country, 0)
			assert.GreaterOrEqual(t, moment.Year, 1990)
			assert.LessOrEqual(t, moment.Year, 2800)
			assert.GreaterOrEqual(t, moment.Month, 1)
			assert.LessOrEqual(t, moment.Month, 12)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			assert.Regexp(t, "[a-zA-Z]+ [0-9]+", moment.Title())
			assert.Regexp(t, "[a-z]+\\-[0-9]+", moment.Slug())
			assert.Regexp(t, "[a-z]+\\-[0-9]+", moment.TitleSlug())
		}
	})
}

func TestMomentsCountries(t *testing.T) {
	t.Run("PublicOnly", func(t *testing.T) {
		results, err := MomentsCountries(1, true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("MomentsCountries %+v", results)

		if len(results) < 1 {
			t.Error("at least one result expected")
		}

		for _, moment := range results {
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
			t.Logf("Title Slug: %s", moment.TitleSlug())

			assert.Len(t, moment.Country, 2)
			assert.GreaterOrEqual(t, moment.Year, 1990)
			assert.LessOrEqual(t, moment.Year, 2800)
			assert.Equal(t, moment.Month, 0)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			assert.Regexp(t, "[ \\&a-zA-Z0-9]+", moment.Title())
			assert.Regexp(t, "[\\-a-z0-9]+", moment.Slug())
			assert.Regexp(t, "[\\-a-z0-9]+", moment.TitleSlug())
		}
	})
	t.Run("IncludePrivate", func(t *testing.T) {
		results, err := MomentsCountries(1, false)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("MomentsCountries %+v", results)

		if len(results) < 1 {
			t.Error("at least one result expected")
		}

		for _, moment := range results {
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
			t.Logf("Title Slug: %s", moment.TitleSlug())

			assert.Len(t, moment.Country, 2)
			assert.GreaterOrEqual(t, moment.Year, 1990)
			assert.LessOrEqual(t, moment.Year, 2800)
			assert.Equal(t, moment.Month, 0)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			assert.Regexp(t, "[ \\&a-zA-Z0-9]+", moment.Title())
			assert.Regexp(t, "[\\-a-z0-9]+", moment.Slug())
			assert.Regexp(t, "[\\-a-z0-9]+", moment.TitleSlug())
		}
	})
}

func TestMomentsStates(t *testing.T) {
	t.Run("PublicOnly", func(t *testing.T) {
		results, err := MomentsStates(1, true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("MomentsStates %+v", results)

		if len(results) < 1 {
			t.Error("at least one result expected")
		}

		for _, moment := range results {
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
			t.Logf("Title Slug: %s", moment.TitleSlug())

			assert.Len(t, moment.Country, 2)
			assert.NotEmpty(t, moment.State)
			assert.Equal(t, moment.Year, 0)
			assert.Equal(t, moment.Month, 0)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			assert.Regexp(t, "[ \\&a-zA-Z0-9]+", moment.Title())
			assert.Regexp(t, "[\\-a-z0-9]+", moment.Slug())
			assert.Regexp(t, "[\\-a-z0-9]+", moment.TitleSlug())
		}
	})
	t.Run("IncludePrivate", func(t *testing.T) {
		results, err := MomentsStates(1, false)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("MomentsStates %+v", results)

		if len(results) < 1 {
			t.Error("at least one result expected")
		}

		for _, moment := range results {
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
			t.Logf("Title Slug: %s", moment.TitleSlug())

			assert.Len(t, moment.Country, 2)
			assert.NotEmpty(t, moment.State)
			assert.Equal(t, moment.Year, 0)
			assert.Equal(t, moment.Month, 0)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			assert.Regexp(t, "[ \\&a-zA-Z0-9]+", moment.Title())
			assert.Regexp(t, "[\\-a-z0-9]+", moment.Slug())
			assert.Regexp(t, "[\\-a-z0-9]+", moment.TitleSlug())
		}
	})
}

func TestMomentsCategories(t *testing.T) {
	t.Run("PublicOnly", func(t *testing.T) {
		results, err := MomentsLabels(1, true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("MomentsLabels %+v", results)

		if len(results) < 1 {
			t.Error("at least one result expected")
		}

		for _, moment := range results {
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
			t.Logf("Title Slug: %s", moment.TitleSlug())

			assert.NotEmpty(t, moment.Label)
			assert.Empty(t, moment.Country)
			assert.Empty(t, moment.State)
			assert.Equal(t, moment.Year, 0)
			assert.Equal(t, moment.Month, 0)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			assert.Regexp(t, "[ \\&a-zA-Z0-9]+", moment.Title())
			assert.Regexp(t, "[\\-a-z0-9]+", moment.Slug())
			assert.Regexp(t, "[\\-a-z0-9]+", moment.TitleSlug())
		}
	})
	t.Run("IncludePrivate", func(t *testing.T) {
		results, err := MomentsLabels(1, false)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("MomentsLabels %+v", results)

		if len(results) < 1 {
			t.Error("at least one result expected")
		}

		for _, moment := range results {
			t.Logf("Title: %s", moment.Title())
			t.Logf("Slug: %s", moment.Slug())
			t.Logf("Title Slug: %s", moment.TitleSlug())

			assert.NotEmpty(t, moment.Label)
			assert.Empty(t, moment.Country)
			assert.Empty(t, moment.State)
			assert.Equal(t, moment.Year, 0)
			assert.Equal(t, moment.Month, 0)
			assert.GreaterOrEqual(t, moment.PhotoCount, 1)
			assert.Regexp(t, "[ \\&a-zA-Z0-9]+", moment.Title())
			assert.Regexp(t, "[\\-a-z0-9]+", moment.Slug())
			assert.Regexp(t, "[\\-a-z0-9]+", moment.TitleSlug())
		}
	})
}

func TestMoment_Title(t *testing.T) {
	t.Run("country", func(t *testing.T) {
		moment := Moment{
			Label:      "",
			Country:    "de",
			State:      "",
			Year:       0,
			Month:      0,
			PhotoCount: 0,
		}

		assert.Equal(t, "Germany", moment.Title())
	})
	t.Run("country name", func(t *testing.T) {
		moment := Moment{
			Label:      "",
			Country:    "de",
			State:      "",
			Year:       1800,
			Month:      0,
			PhotoCount: 0,
		}

		assert.Equal(t, "Germany", moment.Title())
	})
	t.Run("country and year", func(t *testing.T) {
		moment := Moment{
			Label:      "",
			Country:    "de",
			State:      "",
			Year:       2010,
			Month:      0,
			PhotoCount: 0,
		}

		assert.Equal(t, "Germany 2010", moment.Title())
	})
	t.Run("country, state and year", func(t *testing.T) {
		moment := Moment{
			Label:      "",
			Country:    "de",
			State:      "Pfalz",
			Year:       2010,
			Month:      0,
			PhotoCount: 0,
		}

		assert.Equal(t, "Pfalz / 2010", moment.Title())
	})
	t.Run("state, country, month and year", func(t *testing.T) {
		moment := Moment{
			Label:      "",
			Country:    "de",
			State:      "Pfalz",
			Year:       2010,
			Month:      12,
			PhotoCount: 0,
		}

		assert.Equal(t, "Pfalz / December 2010", moment.Title())
	})
	t.Run("month", func(t *testing.T) {
		moment := Moment{
			Label:      "",
			Country:    "",
			State:      "",
			Year:       0,
			Month:      12,
			PhotoCount: 0,
		}

		assert.Equal(t, "December", moment.Title())
	})
}

func TestRemoveDuplicateMoments(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		// Make sure that data is ok before test
		results := Db().Model(entity.PhotoAlbums{}).Where("album_uid NOT IN (?)", Db().Select("album_uid").Model(entity.Album{}).Where("album_uid IS NOT NULL")).Find(&entity.PhotoAlbums{})
		assert.Equal(t, int64(0), results.RowsAffected, "before test orphaned records")

		if removed, err := RemoveDuplicateMoments(); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("moments: removed %s", english.Plural(removed, "duplicate", "duplicates"))

			// This returns a variable number of records on gorm1 due to an issue in sqlite not always deleting 2 records.
			// Updates for gorm2 have not shown this being replicated to date.
			// Delete 2 in Albums and 0 in PhotoAlbums.
			assert.GreaterOrEqual(t, removed, 2)

			results := Db().Model(entity.PhotoAlbums{}).Where("album_uid NOT IN (?)", Db().Select("album_uid").Model(entity.Album{}).Where("album_uid IS NOT NULL")).Find(&entity.PhotoAlbums{})
			assert.Equal(t, int64(0), results.RowsAffected, "after test orphaned records")
		}
	})
}
