package classify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelRules_Find(t *testing.T) {
	result, ok := Rules.Find("cat")
	assert.True(t, ok)
	assert.Equal(t, "cat", result.Label)
	assert.Equal(t, "animal", result.Categories[0])
	assert.Equal(t, 5, result.Priority)
}
