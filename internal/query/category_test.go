package query

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestRepo_CategoryLabels(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.OriginalsPath(), conf.Db())

	categories := search.CategoryLabels(1000, 0)

	assert.Equal(t, "Flower", categories[0].Title)
}
