package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestQuery_LabelBySlug(t *testing.T) {
	conf := config.TestConfig()

	q := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		label, err := q.LabelBySlug("flower")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Flower", label.LabelName)
	})

	t.Run("no files found", func(t *testing.T) {
		label, err := q.LabelBySlug("111")

		assert.Error(t, err, "record not found")
		assert.Empty(t, label.ID)
	})
}

func TestQuery_LabelByUUID(t *testing.T) {
	conf := config.TestConfig()

	q := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		label, err := q.LabelByUUID("lt9k3pw1wowuy3c5")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "COW", label.LabelName)
	})

	t.Run("no files found", func(t *testing.T) {
		label, err := q.LabelByUUID("111")

		assert.Error(t, err, "record not found")
		assert.Empty(t, label.ID)
	})
}

func TestQuery_LabelThumbBySlug(t *testing.T) {
	conf := config.TestConfig()

	q := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		file, err := q.LabelThumbBySlug("flower")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := q.LabelThumbBySlug("cow")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestQuery_LabelThumbByUUID(t *testing.T) {
	conf := config.TestConfig()

	q := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		file, err := q.LabelThumbByUUID("lt9k3pw1wowuy3c4")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := q.LabelThumbByUUID("14")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestQuery_Labels(t *testing.T) {
	conf := config.TestConfig()
	q := New(conf.Db())

	t.Run("search with query", func(t *testing.T) {
		query := form.NewLabelSearch("Query:C Count:1005 Order:slug")
		result, err := q.Labels(query)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("results: %+v", result)

		assert.LessOrEqual(t, 2, len(result))

		for _, r := range result {
			assert.IsType(t, LabelResult{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.LabelName)
			assert.NotEmpty(t, r.LabelSlug)
			assert.NotEmpty(t, r.CustomSlug)

			if fix, ok := entity.LabelFixtures[r.LabelSlug]; ok {
				assert.Equal(t, fix.LabelName, r.LabelName)
				assert.Equal(t, fix.LabelSlug, r.LabelSlug)
				assert.Equal(t, fix.CustomSlug, r.CustomSlug)
			}
		}
	})

	t.Run("search for favorites", func(t *testing.T) {
		query := form.NewLabelSearch("Favorites:true")
		result, err := q.Labels(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(result))

		for _, r := range result {
			assert.True(t, r.LabelFavorite)
			assert.IsType(t, LabelResult{}, r)
			assert.NotEmpty(t, r.ID)
			assert.NotEmpty(t, r.LabelName)
			assert.NotEmpty(t, r.LabelSlug)
			assert.NotEmpty(t, r.CustomSlug)

			if fix, ok := entity.LabelFixtures[r.LabelSlug]; ok {
				assert.Equal(t, fix.LabelName, r.LabelName)
				assert.Equal(t, fix.LabelSlug, r.LabelSlug)
				assert.Equal(t, fix.CustomSlug, r.CustomSlug)
			}
		}
	})

	t.Run("search with empty query", func(t *testing.T) {
		query := form.NewLabelSearch("")
		result, err := q.Labels(query)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 3, len(result))
	})

	t.Run("search with invalid query string", func(t *testing.T) {
		query := form.NewLabelSearch("xxx:bla")
		result, err := q.Labels(query)

		assert.Error(t, err, "unknown filter")
		assert.Empty(t, result)
	})
}
