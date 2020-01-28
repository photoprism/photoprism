package query

import (
	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestRepo_FindLabelBySlug(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.OriginalsPath(), conf.Db())

	t.Run("files found", func(t *testing.T) {
		label, err := search.FindLabelBySlug("flower")

		assert.Nil(t, err)
		assert.Equal(t, "Flower", label.LabelName)
	})

	t.Run("no files found", func(t *testing.T) {
		label, err := search.FindLabelBySlug("111")

		assert.Error(t, err, "record not found")
		t.Log(label)
	})
}

func TestRepo_FindLabelByUUID(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.OriginalsPath(), conf.Db())

	t.Run("files found", func(t *testing.T) {
		label, err := search.FindLabelByUUID("14")

		assert.Nil(t, err)
		assert.Equal(t, "COW", label.LabelName)
	})

	t.Run("no files found", func(t *testing.T) {
		label, err := search.FindLabelByUUID("111")

		assert.Error(t, err, "record not found")
		t.Log(label)
	})
}

func TestRepo_FindLabelThumbBySlug(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.OriginalsPath(), conf.Db())

	t.Run("files found", func(t *testing.T) {
		file, err := search.FindLabelThumbBySlug("flower")

		assert.Nil(t, err)
		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := search.FindLabelThumbBySlug("cow")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestRepo_FindLabelThumbByUUID(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.OriginalsPath(), conf.Db())

	t.Run("files found", func(t *testing.T) {
		file, err := search.FindLabelThumbByUUID("13")

		assert.Nil(t, err)
		assert.Equal(t, "exampleFileName.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := search.FindLabelThumbByUUID("14")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestRepo_Labels(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.OriginalsPath(), conf.Db())

	t.Run("search with query", func(t *testing.T) {
		query := form.NewLabelSearch("Query:C Count:1005 Order:slug")
		result, err := search.Labels(query)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "Cake", result[1].LabelName)
		assert.Equal(t, "COW", result[0].LabelName)
	})

	t.Run("search for favorites", func(t *testing.T) {
		query := form.NewLabelSearch("Favorites:true")
		result, err := search.Labels(query)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "Flower", result[1].LabelName)
		assert.Equal(t, "COW", result[0].LabelName)
	})
}
