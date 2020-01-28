package query

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestRepo_GetMomentsTime(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.OriginalsPath(), conf.Db())

	t.Run("result found", func(t *testing.T) {
		result, err := search.GetMomentsTime()

		assert.Nil(t, err)
		assert.Equal(t, 2790, result[0].PhotoYear)
		assert.Equal(t, 2, result[0].Count)
	})
}
