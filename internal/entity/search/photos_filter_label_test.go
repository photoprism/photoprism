package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
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
		require.Len(t, photos, 3)

		for i := range len(photos) {
			fixture := ""
			switch photos[i].ID {
			case entity.PhotoFixtures.Pointer("19800101_000002_D640C559").ID:
				fixture = "19800101_000002_D640C559"
			case entity.PhotoFixtures.Pointer("Photo56").ID:
				fixture = "Photo56"
			case entity.PhotoFixtures.Pointer("Photo57").ID:
				fixture = "Photo57"
			default:
				assert.Len(t, photos, 0)
				t.Logf("Unable to match id in expected photos %d against %d %d %d", photos[i].ID, entity.PhotoFixtures.Pointer("19800101_000002_D640C559").ID, entity.PhotoFixtures.Pointer("Photo56").ID, entity.PhotoFixtures.Pointer("Photo57").ID)
				t.Fail()
			}
			assert.Equal(t, entity.PhotoFixtures.Pointer(fixture).ID, photos[i].ID)
			assert.Equal(t, entity.PhotoFixtures.Pointer(fixture).CameraID, photos[i].CameraID)
			assert.Equal(t, entity.PhotoFixtures.Pointer(fixture).CellID, photos[i].CellID)
			assert.Equal(t, entity.PhotoFixtures.Pointer(fixture).LensID, photos[i].LensID)
			assert.Equal(t, entity.PhotoFixtures.Pointer(fixture).PhotoCaption, photos[i].PhotoCaption)
		}
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
		var f form.SearchPhotos

		f.Label = "&friendship"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(photos))
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "construction&failure"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(photos))
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Label = "goal&"
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

		f.Query = "label:\"&friendship\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(photos))
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"construction&failure\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(photos))
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "label:\"goal&\""
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
}
