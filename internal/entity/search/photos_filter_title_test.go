package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestPhotosFilterTitle(t *testing.T) {
	t.Run("Lake / 2790", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "Lake / 2790"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("Lake*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "Lake*"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 2)
	})
	t.Run("Neckarbrücke", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "Neckarbrücke"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("Neckarbrücke or Lake*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "Neckarbrücke OR Lake*"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("Neckarbrücke whitespace pipe whitespace Lake*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "Neckarbrücke | Lake*"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 3)
	})
	t.Run("Neckarbrücke pipe Lake*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "Neckarbrücke|Lake*"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 3)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "%Salad"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "hon%ey"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "cloud%"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "&dad"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "leader&ship"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "mom&"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "'poetry'"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "amaz'ing"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "pollution'"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "*area"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "farm*animal"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "engine*"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "|football"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Greater(t, len(photos), 1)
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "pain|ting"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "supermarket|"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 1)
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "123community"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "cli44mate"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "guest456"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "\"member"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "sol\"ution"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "desk\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("OrSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "guest456|engine*"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 2)
	})
	t.Run("OrSearch2", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "guest456|amaz'ing"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 1)
	})
	t.Run("OrSearch3", func(t *testing.T) {
		var f form.SearchPhotos

		f.Title = "farm*animal|amaz'ing"
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 1)
	})
}

func TestPhotosQueryTitle(t *testing.T) {
	t.Run("Lake / 2790", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"Lake / 2790\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("Lake*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"Lake*\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 2)
	})
	t.Run("Neckarbrücke", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"Neckarbrücke\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("Neckarbrücke or Lake*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"Neckarbrücke OR Lake*\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 0)
	})
	t.Run("Neckarbrücke whitespace pipe whitespace Lake*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"Neckarbrücke | Lake*\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 3)
	})
	t.Run("Neckarbrücke pipe Lake*", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"Neckarbrücke|Lake*\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 3)
	})
	t.Run("StartsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"%Salad\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"hon%ey\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithPercent", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"cloud%\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"&dad\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"leader&ship\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithAmpersand", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"mom&\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"'poetry'\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"amaz'ing\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithSingleQuote", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"pollution'\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"*area\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"farm*animal\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithAsterisk", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"engine*\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"|football\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		//TODO
		assert.Equal(t, len(photos), 0)
	})
	t.Run("CenterPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"pain|ting\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(photos))
	})
	t.Run("EndsWithPipe", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"supermarket|\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		//TODO
		assert.Equal(t, len(photos), 0)
	})
	t.Run("StartsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"123community\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 1)
	})
	t.Run("CenterNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"cli44mate\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("EndsWithNumber", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"guest456\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(photos), 1)
	})
	t.Run("StartsWithDoubleQuotes", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"\"member\""
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

		f.Query = "title:\"sol\"ution\""
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

		f.Query = "title:\"desk\"\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		// TODO Finds all?
		assert.Greater(t, len(photos), 1)
	})
	t.Run("OrSearch", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"guest456|engine*\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(photos), 2)
	})
	t.Run("OrSearch2", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"guest456|amaz'ing\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 1)
	})
	t.Run("OrSearch3", func(t *testing.T) {
		var f form.SearchPhotos

		f.Query = "title:\"farm*animal|amaz'ing\""
		f.Merged = true

		photos, _, err := Photos(f)

		if err != nil {
			t.Fatal(err)
		}
		assert.Greater(t, len(photos), 1)
	})
}
