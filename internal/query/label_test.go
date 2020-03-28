package query

import (
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

		assert.Nil(t, err)
		assert.Equal(t, "Flower", label.LabelName)
	})

	t.Run("no files found", func(t *testing.T) {
		label, err := q.LabelBySlug("111")

		assert.Error(t, err, "record not found")
		t.Log(label)
	})
}

func TestQuery_LabelByUUID(t *testing.T) {
	conf := config.TestConfig()

	q := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		label, err := q.LabelByUUID("14")

		assert.Nil(t, err)
		assert.Equal(t, "COW", label.LabelName)
	})

	t.Run("no files found", func(t *testing.T) {
		label, err := q.LabelByUUID("111")

		assert.Error(t, err, "record not found")
		t.Log(label)
	})
}

func TestQuery_LabelThumbBySlug(t *testing.T) {
	conf := config.TestConfig()

	q := New(conf.Db())

	t.Run("files found", func(t *testing.T) {
		file, err := q.LabelThumbBySlug("flower")

		assert.Nil(t, err)
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
		file, err := q.LabelThumbByUUID("13")

		assert.Nil(t, err)
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

		assert.Nil(t, err)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "Cake", result[1].LabelName)
		assert.Equal(t, "COW", result[0].LabelName)
	})

	t.Run("search for favorites", func(t *testing.T) {
		query := form.NewLabelSearch("Favorites:true")
		result, err := q.Labels(query)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "Flower", result[1].LabelName)
		assert.Equal(t, "COW", result[0].LabelName)
	})

	t.Run("search with empty query", func(t *testing.T) {
		query := form.NewLabelSearch("")
		result, err := q.Labels(query)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(result))
	})

	t.Run("search with invalid query string", func(t *testing.T) {
		query := form.NewLabelSearch("xxx:bla")
		result, err := q.Labels(query)
		assert.Error(t, err, "unknown filter")
		t.Log(result)
	})
}
