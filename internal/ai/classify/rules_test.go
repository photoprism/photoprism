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

// TestLabelRulesDogFloor verifies all Dog rules and preserves higher class minimums.
func TestLabelRulesDogFloor(t *testing.T) {
	count := 0
	for name, rule := range Rules {
		if rule.Label != "dog" {
			continue
		}
		count++
		assert.GreaterOrEqual(t, rule.Threshold, float32(0.60), name)
	}
	require.Positive(t, count)
	for name, floor := range map[string]float32{
		"bouvier des flandres":            0.73,
		"bouvier des flandres dog":        0.73,
		"dalmatian":                       0.69,
		"dalmatian dog":                   0.69,
		"dingo":                           0.89,
		"german short-haired pointer":     0.64,
		"german short-haired pointer dog": 0.64,
		"irish water spaniel":             0.66,
		"irish water spaniel dog":         0.66,
		"komondor":                        0.93,
		"komondor dog":                    0.93,
		"schipperke":                      0.995,
		"schipperke dog":                  0.995,
		"sussex spaniel":                  0.6,
		"sussex spaniel dog":              0.6,
		"wire-haired fox terrier":         0.67,
		"wire-haired fox terrier dog":     0.67,
	} {
		require.Contains(t, Rules, name)
		assert.Equal(t, floor, Rules[name].Threshold, name)
	}
}
