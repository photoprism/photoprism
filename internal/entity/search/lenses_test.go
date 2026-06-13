package search

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

func TestLenses(t *testing.T) {
	t.Run("SearchWithQuery", func(t *testing.T) {
		query := form.NewLensSearch("q:A")
		query.Count = 1005
		result, err := Lenses(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(result))

		for _, r := range result {
			assert.IsType(t, Lens{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.LensName)
			assert.NotEmpty(t, r.LensSlug)
			assert.NotEmpty(t, r.LensModel)
			assert.NotEmpty(t, r.LensMake)
			assert.True(t, strings.Contains(strings.ToLower(r.LensName), "a") || strings.Contains(strings.ToLower(r.LensMake), "a") || strings.Contains(strings.ToLower(r.LensModel), "a"))

			if fix, ok := entity.LensFixtures[r.LensSlug]; ok {
				assert.Equal(t, fix.LensName, r.LensName)
				assert.Equal(t, fix.LensSlug, r.LensSlug)
				assert.Equal(t, fix.LensMake, r.LensMake)
				assert.Equal(t, fix.LensModel, r.LensModel)
			}
		}
	})
	t.Run("SearchForString", func(t *testing.T) {
		query := form.NewLensSearch("Q:4.15mm")
		query.Count = 1005
		result, err := Lenses(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(result))

		for _, r := range result {
			assert.IsType(t, Lens{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.LensName)
			assert.NotEmpty(t, r.LensSlug)
			assert.NotEmpty(t, r.LensModel)

			if fix, ok := entity.LensFixtures[r.LensSlug]; ok {
				assert.Equal(t, fix.LensName, r.LensName)
				assert.Equal(t, fix.LensSlug, r.LensSlug)
				assert.Equal(t, fix.LensMake, r.LensMake)
				assert.Equal(t, fix.LensModel, r.LensModel)
			}
		}
	})
	t.Run("SearchForNoMake", func(t *testing.T) {
		query := form.NewLensSearch("NoMake:true")
		query.Count = 15
		result, err := Lenses(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(result))

		for _, r := range result {
			assert.IsType(t, Lens{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.LensName)
			assert.NotEmpty(t, r.LensSlug)
			assert.Empty(t, r.LensMake)

			if fix, ok := entity.LensFixtures[r.LensSlug]; ok {
				assert.Equal(t, fix.LensName, r.LensName)
				assert.Equal(t, fix.LensSlug, r.LensSlug)
				assert.Equal(t, fix.LensModel, r.LensModel)
			}
		}
	})
	t.Run("SearchWithEmptyQuery", func(t *testing.T) {
		fixture := "4-37"
		query := form.NewLensSearch("")
		result, err := Lenses(query)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 3, len(result))
		assert.Equal(t, entity.LensFixtures.Get(fixture).LensSlug, result[0].LensSlug)
	})
	t.Run("SearchWithReverse", func(t *testing.T) {
		fixture := "4.15mm-f/2.2"
		query := form.NewLensSearch("")
		query.Reverse = true
		result, err := Lenses(query)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 3, len(result))
		assert.Equal(t, entity.LensFixtures.Get(fixture).LensSlug, result[0].LensSlug)
	})
	t.Run("SearchForNoMakeAndQueryNoResults", func(t *testing.T) {
		query := form.NewLensSearch("Apple")
		query.NoMake = true
		result, err := Lenses(query)

		assert.NoError(t, err)
		assert.Empty(t, result)
	})
	t.Run("SearchWithInvalidQueryString", func(t *testing.T) {
		query := form.NewLensSearch("xxx:bla")
		result, err := Lenses(query)

		assert.Error(t, err, "unknown filter")
		assert.Empty(t, result)
	})
	t.Run("SearchForId", func(t *testing.T) {
		f := form.SearchLenses{
			Query:   "",
			ID:      "1000002|1000000",
			Slug:    "",
			Name:    "",
			NoMake:  false,
			Count:   0,
			Offset:  0,
			Reverse: false,
		}

		result, err := Lenses(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, result, 2)
		assert.Equal(t, "4-37", result[0].LensSlug)
	})
}
