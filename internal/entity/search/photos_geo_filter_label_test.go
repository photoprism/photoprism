package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

// geoBaselineCount returns the merged geo-result count for a search with no
// label filter.
func geoBaselineCount(t *testing.T) int {
	t.Helper()

	result, err := PhotosGeo(form.NewSearchPhotosGeo(""))
	if err != nil {
		t.Fatal(err)
	}

	return len(result)
}

// geoCountForLabel runs a geo search with the given label filter and returns
// the number of results.
func geoCountForLabel(t *testing.T, label string) int {
	t.Helper()

	q := form.NewSearchPhotosGeo("")
	q.Label = label

	result, err := PhotosGeo(q)
	if err != nil {
		t.Fatal(err)
	}

	return len(result)
}

func TestPhotosGeoFilterLabel(t *testing.T) {
	t.Run("SingleInclude", func(t *testing.T) {
		assert.Greater(t, geoCountForLabel(t, "cake"), 0)
	})
	t.Run("SingleExclude", func(t *testing.T) {
		base := geoBaselineCount(t)
		withFlower := geoCountForLabel(t, "flower")
		result := geoCountForLabel(t, "!flower")
		assert.Equal(t, base-withFlower, result)
	})
	t.Run("ExcludeUnknownIsNoOp", func(t *testing.T) {
		base := geoBaselineCount(t)
		result := geoCountForLabel(t, "!totally-unknown-label")
		assert.Equal(t, base, result)
	})
	t.Run("OnlyNegative", func(t *testing.T) {
		base := geoBaselineCount(t)
		withCake := geoCountForLabel(t, "cake")
		result := geoCountForLabel(t, "!cake")
		assert.Equal(t, base-withCake, result)
	})
	t.Run("IncludeAndExclude", func(t *testing.T) {
		// cake&!flower is empty on the geo subset because every geotagged
		// cake photo is also tagged with flower.
		result := geoCountForLabel(t, "cake&!flower")
		assert.Equal(t, 0, result)
	})
	t.Run("MultipleIncludes", func(t *testing.T) {
		result := geoCountForLabel(t, "cake&flower")
		assert.Equal(t, 2, result)
	})
	t.Run("MultipleIncludesPlusExclude", func(t *testing.T) {
		result := geoCountForLabel(t, "cake&flower&!cow")
		assert.Equal(t, 0, result)
	})
	t.Run("IncludeOrExcludeCow", func(t *testing.T) {
		result := geoCountForLabel(t, "cake|flower&!cow")
		// All geotagged cake-or-flower photos also carry cow in fixtures.
		assert.Equal(t, 0, result)
	})
	t.Run("CategoryExpansion", func(t *testing.T) {
		base := geoBaselineCount(t)
		withLandscape := geoCountForLabel(t, "landscape")
		assert.Greater(t, withLandscape, 0)
		assert.Equal(t, base-withLandscape, geoCountForLabel(t, "!landscape"))
	})
	t.Run("EscapeLiteralBang", func(t *testing.T) {
		// No fixture label named "!weird" exists; positive lookup short-circuits.
		assert.Equal(t, 0, geoCountForLabel(t, `\!weird`))
	})
	t.Run("EscapeLiteralBangNegated", func(t *testing.T) {
		base := geoBaselineCount(t)
		assert.Equal(t, base, geoCountForLabel(t, `!\!weird`))
	})
	t.Run("LegacyAmpersandName", func(t *testing.T) {
		assert.Equal(t, 1, geoCountForLabel(t, `construction\&failure`))
		// Unescaped now parses as two positive AND groups, neither resolvable.
		assert.Equal(t, 0, geoCountForLabel(t, "construction&failure"))
	})
}

func TestPhotosGeoQueryLabel(t *testing.T) {
	t.Run("SingleExclude", func(t *testing.T) {
		base := geoBaselineCount(t)
		withFlower := geoCountForLabel(t, "flower")

		q := form.NewSearchPhotosGeo(`label:"!flower"`)
		result, err := PhotosGeo(q)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, base-withFlower, len(result))
	})
	t.Run("IncludeAndExclude", func(t *testing.T) {
		q := form.NewSearchPhotosGeo(`label:"cake&flower"`)
		result, err := PhotosGeo(q)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(result))
	})
}
