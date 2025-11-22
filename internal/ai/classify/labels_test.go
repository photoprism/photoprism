package classify

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabel_AppendLabel(t *testing.T) {
	cat := Label{Name: "cat", Source: "location", Uncertainty: 80, Priority: 5}
	dog := Label{Name: "dog", Source: "location", Uncertainty: 80, Priority: 5}
	labels := Labels{cat, dog}

	t.Run("LabelWithName", func(t *testing.T) {
		assert.Equal(t, 2, labels.Len())
		cow := Label{Name: "cow", Source: "location", Uncertainty: 80, Priority: 5}
		labelsNew := labels.AppendLabel(cow)
		assert.Equal(t, 3, labelsNew.Len())
		assert.Equal(t, "dog", labelsNew[1].Name)
		assert.Equal(t, "cat", labelsNew[0].Name)
		assert.Equal(t, "cow", labelsNew[2].Name)
	})
	t.Run("LabelWithoutName", func(t *testing.T) {
		assert.Equal(t, 2, labels.Len())
		cow := Label{Name: "", Source: "location", Uncertainty: 80, Priority: 5}
		labelsNew := labels.AppendLabel(cow)
		assert.Equal(t, 2, labelsNew.Len())
		assert.Equal(t, "dog", labelsNew[1].Name)
	})
}

func TestLabels_Title(t *testing.T) {
	t.Run("First", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 59, Priority: 5}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 10, Priority: 4}
		labels := Labels{cat, dog}

		assert.Equal(t, "cat", labels.Title("fallback"))
	})
	t.Run("Second", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 61, Priority: 5}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 10, Priority: 4}
		labels := Labels{cat, dog}

		assert.Equal(t, "dog", labels.Title("fallback"))
	})
	t.Run("Fallback", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 80, Priority: 5}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 80, Priority: 4}
		labels := Labels{cat, dog}

		assert.Equal(t, "fallback", labels.Title("fallback"))
	})
	t.Run("EmptyLabels", func(t *testing.T) {
		labels := Labels{}

		assert.Equal(t, "", labels.Title(""))
	})
	t.Run("LabelPriorityLessThanZero", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 59, Priority: -1}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 10, Priority: -1}
		labels := Labels{cat, dog}

		assert.Equal(t, "fallback", labels.Title("fallback"))
	})
	t.Run("LabelPriorityEqualZero", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 59, Priority: 0}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 62, Priority: 0}
		labels := Labels{cat, dog}

		assert.Equal(t, "fallback", labels.Title("fallback"))
	})
}

func TestLabels_Keywords(t *testing.T) {
	cat := Label{Name: "cat", Source: "location", Uncertainty: 80, Priority: 5, Categories: []string{"animal"}}
	dog := Label{Name: "dog", Source: "location", Uncertainty: 80, Priority: 5}
	bird := Label{Name: "bird", Source: "image", Uncertainty: 100, Priority: 2}
	labels := Labels{cat, dog, bird}

	t.Run("LabelWithName", func(t *testing.T) {
		result := labels.Keywords()
		t.Log(result)
		assert.Equal(t, "cat", result[0])
		assert.Equal(t, "animal", result[1])
		assert.Equal(t, "dog", result[2])
		assert.Equal(t, 3, len(result))
	})
}

func TestLabels_Names(t *testing.T) {
	t.Run("FiltersEmptyAndUncertain", func(t *testing.T) {
		labels := Labels{
			{Name: "cat", Uncertainty: 20},
			{Name: "", Uncertainty: 10},
			{Name: "dog", Uncertainty: 100},
			{Name: "bird", Uncertainty: 99},
		}

		assert.Equal(t, []string{"cat", "bird"}, labels.Names())
	})

	t.Run("NilLabels", func(t *testing.T) {
		var labels Labels
		assert.Nil(t, labels.Names())
	})
}

func TestLabels_Count(t *testing.T) {
	t.Run("CountsEligible", func(t *testing.T) {
		labels := Labels{
			{Name: "cat", Uncertainty: 20},
			{Name: "", Uncertainty: 10},
			{Name: "dog", Uncertainty: 100},
			{Name: "bird", Uncertainty: 99},
		}

		assert.Equal(t, 2, labels.Count())
	})

	t.Run("NilLabels", func(t *testing.T) {
		var labels Labels
		assert.Equal(t, 0, labels.Count())
	})
}

func TestLabels_String(t *testing.T) {
	t.Run("JoinWithAnd", func(t *testing.T) {
		labels := Labels{{Name: "cat"}, {Name: "dog"}, {Name: "bird"}}
		assert.Equal(t, "cat, dog, and bird", labels.String())
	})

	t.Run("NoneForNil", func(t *testing.T) {
		var labels Labels
		assert.Equal(t, "none", labels.String())
	})
}

func TestLabels_IsNSFW(t *testing.T) {
	cases := []struct {
		name      string
		labels    Labels
		threshold int
		expected  bool
	}{
		{
			name:      "ExplicitFlag",
			threshold: 80,
			labels:    Labels{{Name: "cat", NSFW: true}},
			expected:  true,
		},
		{
			name:      "ConfidenceAboveThreshold",
			threshold: 80,
			labels:    Labels{{Name: "cat", NSFWConfidence: 85}},
			expected:  true,
		},
		{
			name:      "ThresholdClamped",
			threshold: 150,
			labels:    Labels{{Name: "cat", NSFWConfidence: 90}},
			expected:  false,
		},
		{
			name:      "BelowThreshold",
			threshold: 80,
			labels:    Labels{{Name: "cat", NSFWConfidence: 40}},
			expected:  false,
		},
		{
			name:      "NegativeThreshold",
			threshold: -10,
			labels:    Labels{{Name: "cat", NSFWConfidence: 100}},
			expected:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.labels.IsNSFW(tc.threshold))
		})
	}
}

func TestLabel_Sort(t *testing.T) {
	labels := Labels{
		{Name: "label 0", Source: "location", Uncertainty: 100, Priority: 10},
		{Name: "label 1", Source: "location", Uncertainty: 100, Priority: -1},
		{Name: "label 2", Source: "location", Uncertainty: 80, Priority: 5},
		{Name: "label 3", Source: "location", Uncertainty: 80, Priority: 5},
		{Name: "label 4", Source: "location", Uncertainty: 99, Priority: 5},
		{Name: "label 5", Source: "location", Uncertainty: 1, Priority: 0},
		{Name: "label 6", Source: "location", Uncertainty: 0, Priority: 5},
		{Name: "label 7", Source: "location", Uncertainty: 0, Priority: 1},
		{Name: "label 8", Source: "location", Uncertainty: 101, Priority: 5},
	}

	sort.Sort(labels)

	assert.Equal(t, "label 6", labels[0].Name)
	assert.Equal(t, "label 2", labels[1].Name)
	assert.Equal(t, "label 3", labels[2].Name)
	assert.Equal(t, "label 4", labels[3].Name)
	assert.Equal(t, "label 7", labels[4].Name)
	assert.Equal(t, "label 5", labels[5].Name)
	assert.Equal(t, "label 0", labels[6].Name)
	assert.Equal(t, "label 1", labels[7].Name)
	assert.Equal(t, "label 8", labels[8].Name)
}
