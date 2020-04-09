package query

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestQuery_CategoryLabels(t *testing.T) {
	conf := config.TestConfig()

	search := New(conf.Db())

	categories := search.CategoryLabels(1000, 0)

	assert.GreaterOrEqual(t, 1, len(categories))
}
