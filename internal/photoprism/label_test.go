package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabel_NewLocationLabel(t *testing.T) {
	LocLabel := NewLocationLabel("locationtest", 23, 1)
	t.Log(LocLabel)
	assert.Equal(t, "location", LocLabel.Source)
	assert.Equal(t, 23, LocLabel.Uncertainty)
	assert.Equal(t, "locationtest", LocLabel.Name)

	t.Run("locationtest / slash", func(t *testing.T) {
		LocLabel := NewLocationLabel("locationtest / slash", 24, -2)
		t.Log(LocLabel)
		assert.Equal(t, "location", LocLabel.Source)
		assert.Equal(t, 24, LocLabel.Uncertainty)
		assert.Equal(t, "locationtest", LocLabel.Name)
	})

	t.Run("locationtest - minus", func(t *testing.T) {
		LocLabel := NewLocationLabel("locationtest - minus", 80, -2)
		t.Log(LocLabel)
		assert.Equal(t, "location", LocLabel.Source)
		assert.Equal(t, 80, LocLabel.Uncertainty)
		assert.Equal(t, "locationtest", LocLabel.Name)
	})
}

func TestLabel_AppendLabel(t *testing.T) {
	cat := Label{Name: "cat", Source: "location", Uncertainty: 80, Priority: 5}
	dog := Label{Name: "dog", Source: "location", Uncertainty: 80, Priority: 5}
	labels := Labels{cat, dog}

	t.Run("labelWithName", func(t *testing.T) {
		assert.Equal(t, 2, labels.Len())
		cow := Label{Name: "cow", Source: "location", Uncertainty: 80, Priority: 5}
		labelsNew := labels.AppendLabel(cow)
		assert.Equal(t, 3, labelsNew.Len())
		assert.Equal(t, "dog", labelsNew[1].Name)
		assert.Equal(t, "cat", labelsNew[0].Name)
		assert.Equal(t, "cow", labelsNew[2].Name)
	})

	t.Run("labelWithoutName", func(t *testing.T) {
		assert.Equal(t, 2, labels.Len())
		cow := Label{Name: "", Source: "location", Uncertainty: 80, Priority: 5}
		labelsNew := labels.AppendLabel(cow)
		assert.Equal(t, 2, labelsNew.Len())
		assert.Equal(t, "dog", labelsNew[1].Name)
	})

}

func TestLabels_Title(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 59, Priority: 5}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 10, Priority: 4}
		labels := Labels{cat, dog}

		assert.Equal(t, "cat", labels.Title("fallback"))
	})

	t.Run("second", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 61, Priority: 5}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 10, Priority: 4}
		labels := Labels{cat, dog}

		assert.Equal(t, "dog", labels.Title("fallback"))
	})

	t.Run("fallback", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 80, Priority: 5}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 80, Priority: 4}
		labels := Labels{cat, dog}

		assert.Equal(t, "fallback", labels.Title("fallback"))
	})

	t.Run("empty fallback", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 80, Priority: 5}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 80, Priority: 4}
		labels := Labels{cat, dog}

		assert.Equal(t, "", labels.Title(""))
	})

	t.Run("empty labels", func(t *testing.T) {
		labels := Labels{}

		assert.Equal(t, "fallback", labels.Title("fallback"))
	})

	t.Run("priority < 0", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 61, Priority: -5}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 10, Priority: -4}
		labels := Labels{cat, dog}

		assert.Equal(t, "fallback", labels.Title("fallback"))
	})

	t.Run("priority == 0", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 60, Priority: 0}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 51, Priority: 0}
		labels := Labels{cat, dog}

		assert.Equal(t, "fallback", labels.Title("fallback"))
	})
}

func TestLabels_Len(t *testing.T) {
	t.Run("len = 2", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 59, Priority: 5}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 10, Priority: 4}
		labels := Labels{cat, dog}
		assert.Equal(t, 2, labels.Len())
	})
}

func TestLabels_Swap(t *testing.T) {
	t.Run("swap cat with dog", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 59, Priority: 5}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 10, Priority: 4}
		labels := Labels{cat, dog}
		assert.Equal(t, "dog", labels[1].Name)
		labels.Swap(0, 1)
		assert.Equal(t, "cat", labels[1].Name)
	})
}

func TestLabels_Less(t *testing.T) {
	t.Run("different priorities", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 59, Priority: 5}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 10, Priority: 4}
		labels := Labels{cat, dog}
		assert.Equal(t, "dog", labels[1].Name)
		assert.True(t, labels.Less(0, 1))
		assert.False(t, labels.Less(1, 0))
	})

	t.Run("equal priorities", func(t *testing.T) {
		cat := Label{Name: "cat", Source: "location", Uncertainty: 59, Priority: 5}
		dog := Label{Name: "dog", Source: "location", Uncertainty: 10, Priority: 5}
		labels := Labels{cat, dog}
		assert.Equal(t, "dog", labels[1].Name)
		assert.False(t, labels.Less(0, 1))
		assert.True(t, labels.Less(1, 0))
	})
}

func TestLabels_Keywords(t *testing.T) {
	cat := Label{Name: "cat", Source: "location", Uncertainty: 59, Priority: 5, Categories: []string{"animal"}}
	dog := Label{Name: "dog", Source: "location", Uncertainty: 10, Priority: 5, Categories: []string{"animal"}}
	labels := Labels{cat, dog}
	result := labels.Keywords()
	assert.Equal(t, "cat", result[0])
	assert.Equal(t, "animal", result[1])
	assert.Equal(t, "dog", result[2])
	assert.Equal(t, "animal", result[3])

}
