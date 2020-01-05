package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestRepo_CategoryLabels(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.OriginalsPath(), conf.Db())

	categories := search.CategoryLabels(1000, 0)

	t.Logf("categories: %+v", categories)
}
