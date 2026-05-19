package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestPhotosFilterLabel(t *testing.T) {
	t.Run("Flower", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "flower"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 3)
	})
	t.Run("Cake", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "cake"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 5)
	})
	t.Run("CakePipeFlower", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "cake|flower"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 5)
	})
	t.Run("CakeWhitespacePipeWhitespaceFlower", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "cake | flower"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 5)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "%tennis"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 1)
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "chem%stry"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 1)
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "cell%"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		// Option A: a literal '&' inside a label name must be escaped because
		// the unescaped form is now parsed as an AND separator between groups.
		var f form.SearchPhotos

		f.Label = `\&friendship`
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(photos))
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = `construction\&failure`
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(photos))
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = `goal\&`
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "'activity"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "funera'l"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "technology'"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "*tea"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "soup*menu"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "proposal*"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "|college"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "potato|couch"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "mall|"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "2020-world"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 1)
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "sport-2021-event"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "oven-3000"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("StartsWithDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "\"king"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("CenterDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "sal\"mon"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 0)
	})
	t.Run("EndsWithDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "ladder\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 4)
	})
	t.Run("OrSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "oven-3000 | sport-2021-event"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 2)
	})
}

func TestPhotosQueryLabel(t *testing.T) {
	t.Run("Flower", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"flower\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 3)
	})
	t.Run("Cake", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"cake\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 5)
	})
	t.Run("CakePipeFlower", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"cake|flower\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 5)
	})
	t.Run("CakeWhitespacePipeWhitespaceFlower", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"cake | flower\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 5)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"%tennis\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 1)
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"chem%stry\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 1)
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"cell%\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(photos))
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = `label:"\&friendship"`
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(photos))
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = `label:"construction\&failure"`
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(photos))
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = `label:"goal\&"`
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"'activity\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"funera'l\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"technology'\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"*tea\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"soup*menu\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"proposal*\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"|college\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"potato|couch\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"mall|\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"2020-world\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 1)
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"sport-2021-event\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"oven-3000\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, photos, 1)
	})
	t.Run("StartsWithDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"\"king\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		// TODO Finds all?
		assert.Greater(t, len(photos), 0)
	})
	t.Run("CenterDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"sal\"mon\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		// TODO Finds all?
		assert.Greater(t, len(photos), 0)
	})
	t.Run("EndsWithDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"ladder\"\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		// TODO Finds all?
		assert.Greater(t, len(photos), 0)
	})
	t.Run("OrSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"oven-3000 | sport-2021-event\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, photos, 2)
	})
	t.Run("HomophoneName", func(t *testing.T) {
		_, second, _, photoB := createHomophoneSearchFixtures(t)

		var f form.SearchPhotos

		f.Label = second.LabelName
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		if len(photos) != 1 {
			t.Fatalf("expected one result, got %d", len(photos))
		}

		assert.Equal(t, photoB.PhotoUID, photos[0].PhotoUID)
	})
}

// baselinePhotoCount returns the merged-photo result count for a search with no
// label filter. Used by NOT/AND tests to derive expected counts.
func baselinePhotoCount(t *testing.T) int {
	t.Helper()

	var f form.SearchPhotos

	f.Merged = true

	photos, _, err := Photos(f)
	if err != nil {
		t.Fatal(err)
	}

	return len(photos)
}

// photosWithLabel runs a label filter search and returns the merged photos.
func photosWithLabel(t *testing.T, label string) PhotoResults {
	t.Helper()

	var f form.SearchPhotos

	f.Label = label
	f.Merged = true

	photos, _, err := Photos(f)
	if err != nil {
		t.Fatal(err)
	}

	return photos
}

// containsPhotoUID reports whether any of the photos in results carries the
// given UID.
func containsPhotoUID(results PhotoResults, uid string) bool {
	for _, p := range results {
		if p.PhotoUID == uid {
			return true
		}
	}

	return false
}

func TestPhotosFilterLabelNotAnd(t *testing.T) {
	t.Run("SingleExclude", func(t *testing.T) {
		base := baselinePhotoCount(t)
		withFlower := len(photosWithLabel(t, "flower"))
		result := photosWithLabel(t, "!flower")
		assert.Equal(t, base-withFlower, len(result))
	})
	t.Run("ExcludeUnknownIsNoOp", func(t *testing.T) {
		base := baselinePhotoCount(t)
		result := photosWithLabel(t, "!totally-unknown-label")
		assert.Equal(t, base, len(result))
	})
	t.Run("OnlyNegative", func(t *testing.T) {
		base := baselinePhotoCount(t)
		withCake := len(photosWithLabel(t, "cake"))
		result := photosWithLabel(t, "!cake")
		assert.Equal(t, base-withCake, len(result))
	})
	t.Run("IncludeAndExclude", func(t *testing.T) {
		result := photosWithLabel(t, "cake&!flower")
		assert.Len(t, result, 2)
	})
	t.Run("MultipleIncludes", func(t *testing.T) {
		result := photosWithLabel(t, "cake&flower")
		assert.Len(t, result, 3)
	})
	t.Run("MultipleIncludesPlusExclude", func(t *testing.T) {
		result := photosWithLabel(t, "cake&flower&!cow")
		assert.Len(t, result, 1)
	})
	t.Run("IncludeOrExcludeBlurry", func(t *testing.T) {
		result := photosWithLabel(t, "cake|flower&!cow")
		assert.Len(t, result, 2)
	})
	t.Run("CategoryExpansion", func(t *testing.T) {
		base := baselinePhotoCount(t)
		withLandscape := len(photosWithLabel(t, "landscape"))
		result := photosWithLabel(t, "!landscape")
		assert.Equal(t, base-withLandscape, len(result))
		assert.Greater(t, withLandscape, 0)
	})
	t.Run("EscapeLiteralBang", func(t *testing.T) {
		// No fixture label named "!weird" exists, so a positive lookup for the
		// escaped literal short-circuits to an empty result set.
		result := photosWithLabel(t, `\!weird`)
		assert.Empty(t, result)
	})
	t.Run("EscapeLiteralBangNegated", func(t *testing.T) {
		// Negating an unknown literal label "!weird" is a no-op.
		base := baselinePhotoCount(t)
		result := photosWithLabel(t, `!\!weird`)
		assert.Equal(t, base, len(result))
	})
	t.Run("LegacyAmpersandName", func(t *testing.T) {
		escaped := photosWithLabel(t, `construction\&failure`)
		assert.Len(t, escaped, 2)

		// The same input without escape is now parsed as two positive AND
		// groups: neither "construction" nor "failure" exists as a fixture
		// label, so the result short-circuits to empty.
		unescaped := photosWithLabel(t, "construction&failure")
		assert.Empty(t, unescaped)
	})
	t.Run("OnlyNegativeTwoGroups", func(t *testing.T) {
		base := baselinePhotoCount(t)
		// With disjoint label sets, the count equals base minus union; since
		// flower ⊂ landscape-category here, use an independent pair.
		withCake := len(photosWithLabel(t, "cake"))
		withCow := len(photosWithLabel(t, "cow"))
		result := photosWithLabel(t, "!cake&!cow")
		// The difference must be between base-(cake+cow) and base-max(cake,cow).
		maxSide := withCake
		if withCow > maxSide {
			maxSide = withCow
		}

		assert.LessOrEqual(t, len(result), base-maxSide)
		assert.GreaterOrEqual(t, len(result), base-(withCake+withCow))
	})
}
