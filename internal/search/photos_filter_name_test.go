package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestPhotosFilterName(t *testing.T) {
	t.Run("27900704_070228_D6D51B6C", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "27900704_070228_D6D51B6C"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("*hoto1*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "*hoto1*"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 6)
	})
	t.Run("*hoto1* pipe 27900704_070228_D6D51B6C", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "*hoto1*|27900704_070228_D6D51B6C"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 7)
	})
	t.Run("*hoto1* whitespace pipe whitespace 27900704_070228_D6D51B6C", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "*hoto1* | 27900704_070228_D6D51B6C"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 7)
	})
	t.Run("*hoto1* or 27900704_070228_D6D51B6C", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "*hoto1* or 27900704_070228_D6D51B6C"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo29%"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo%30"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo29%"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "&photo31"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo&32"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo33&"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "'photo34"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo'35"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo36'"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "*photo37"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo*38"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo39*"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithPipeWildcard", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "|*photo40"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterPipeWildcard", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo|*41"
		f.Merged = true

		photos, count, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		if len(photos) != 1 {
			t.Logf("excactly one result expected, but %d photos with %d files found", len(photos), count)
			t.Logf("query results: %#v", photos)
		}

		assert.Equal(t, 1, len(photos))
	})
	t.Run("EndsWithPipeWildcard", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo42*|"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "|photo40"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo|41"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo42|"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "43photo"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "pho44to"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo45"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "\"photo46"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo\"47"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo48\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithWhitespace", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = " photo49"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterWhitespace", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo 50"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithWhitespace", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo51 "
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("OrSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "%photo28 | photo'35"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 2)
	})
	t.Run("OrSearch2", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo*38 | photo'35"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 1)
	})
	t.Run("OrSearch3", func(t *testing.T) {
		var f form.SearchPhotos

		f.Name = "photo|41 | &photo31"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 1, len(photos))
	})
}

func TestPhotosQueryName(t *testing.T) {
	t.Run("27900704_070228_D6D51B6C", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"27900704_070228_D6D51B6C\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("*hoto1*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"*hoto1*\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 6)
	})
	t.Run("*hoto1* pipe 27900704_070228_D6D51B6C", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"*hoto1*|27900704_070228_D6D51B6C\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 7)
	})
	t.Run("*hoto1* whitespace pipe whitespace 27900704_070228_D6D51B6C", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"*hoto1* | 27900704_070228_D6D51B6C\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 7)
	})
	t.Run("*hoto1* or 27900704_070228_D6D51B6C", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"*hoto1* or 27900704_070228_D6D51B6C\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(photos))
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"%photo28\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo%30\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo29%\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"&photo31\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo&32\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo33&\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"'photo34\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo'35\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo36'\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"*photo37\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo*38\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo39*\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithPipeWildcard", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"|*photo40\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterPipeWildcard", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo|*41\""
		f.Merged = true

		photos, count, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		if len(photos) != 1 {
			t.Logf("excactly one result expected, but %d photos with %d files found", len(photos), count)
			t.Logf("query results: %#v", photos)
		}

		assert.Equal(t, 1, len(photos))
	})
	t.Run("EndsWithPipeWildcard", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo42*|\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"|photo40\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo|41\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo42|\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"43photo\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"pho44to\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo45\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"\"photo46\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		// TODO Finds all?
		assert.Greater(t, len(photos), 1)
	})
	t.Run("CenterDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo\"47\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		// TODO Finds all?
		assert.Greater(t, len(photos), 1)
	})
	t.Run("EndsWithDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo48\"\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		// TODO Finds all?
		assert.Greater(t, len(photos), 1)
	})
	t.Run("StartsWithWhitespace", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\" photo49\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterWhitespace", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo 50\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithWhitespace", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo51 \""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("OrSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"%photo28 | photo'35\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 2)
	})
	t.Run("OrSearch2", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo*38 | photo'35\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 1)
	})
	t.Run("OrSearch3", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"photo|41 | &photo31\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 1, len(photos))
	})
	t.Run("OrSearch4", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "name:\"bridge1 | bridge2\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 2, len(photos))
	})
}
