package classify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelRule_Find(t *testing.T) {
	var rules = LabelRules{
		"cat": {
			Label:      "",
			Threshold:  1.000000,
			Priority:   -2,
			Categories: []string{"animal"},
		},
		"dog": {
			Label:      "portrait",
			Threshold:  0.200000,
			Priority:   0,
			Categories: []string{"people"},
		},
	}

	t.Run("existing rule", func(t *testing.T) {
		result := rules.Find("cat")
		assert.Equal(t, -2, result.Priority)
		assert.Equal(t, float32(1), result.Threshold)
	})

	t.Run("not existing rule", func(t *testing.T) {
		result := rules.Find("fish")
		t.Log(result)
		assert.Equal(t, float32(0.1), result.Threshold)
	})
}
