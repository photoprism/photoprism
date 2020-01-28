package query

import (
	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestRepo_Geo(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.OriginalsPath(), conf.Db())

	t.Run("search all photos", func(t *testing.T) {
		query := form.NewGeoSearch("")
		result, err := search.Geo(query)

		assert.Nil(t, err)
		assert.Equal(t, 4, len(result))

	})

	t.Run("search for bridge", func(t *testing.T) {
		query := form.NewGeoSearch("Query:bridge Before:3006-01-02")
		result, err := search.Geo(query)

		assert.Nil(t, err)
		assert.Equal(t, "Neckarbrücke", result[0].PhotoTitle)

	})

	t.Run("search for timeframe", func(t *testing.T) {
		query := form.NewGeoSearch("After:2014-12-02 Before:3006-01-02")
		result, err := search.Geo(query)

		assert.Nil(t, err)
		t.Log(result)
		assert.Equal(t, "Reunion", result[0].PhotoTitle)

	})
}
