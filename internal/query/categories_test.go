package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategoryLabels(t *testing.T) {
	categories := CategoryLabels(1000, 0)

	assert.GreaterOrEqual(t, 1, len(categories))
}
