package classify

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// TestLabelRules_Find verifies an existing rule retains its mapping and priority.
func TestLabelRules_Find(t *testing.T) {
	result, ok := Rules.Find("cat")
	assert.True(t, ok)
	assert.Equal(t, "cat", result.Label)
	assert.Equal(t, "animal", result.Categories[0])
	assert.Equal(t, 5, result.Priority)
}

// TestLabelRulesSource verifies unique source keys and complete generated alias resolution.
func TestLabelRulesSource(t *testing.T) {
	var source map[string]struct {
		LabelRule `yaml:",inline"`
		See       string
	}
	data, err := os.ReadFile("rules.yml")
	require.NoError(t, err)
	require.NoError(t, yaml.UnmarshalStrict(data, &source))
	require.Len(t, Rules, len(source))

	for name, entry := range source {
		require.Equal(t, strings.ToLower(name), name)
		if entry.See != "" {
			target, ok := source[entry.See]
			require.True(t, ok, name)
			require.Empty(t, target.See, name)
			entry.LabelRule = target.LabelRule
		}
		if entry.Categories == nil {
			entry.Categories = []string{}
		}
		require.Equal(t, entry.LabelRule, Rules[name], name)
	}
}
