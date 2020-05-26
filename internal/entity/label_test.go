package entity

import (
	"github.com/photoprism/photoprism/internal/classify"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLabel(t *testing.T) {
	t.Run("name Unicorn2000 priority 5", func(t *testing.T) {
		label := NewLabel("Unicorn2000", 5)
		assert.Equal(t, "Unicorn2000", label.LabelName)
		assert.Equal(t, "unicorn2000", label.LabelSlug)
		assert.Equal(t, 5, label.LabelPriority)
	})
	t.Run("name Unknown", func(t *testing.T) {
		label := NewLabel("", -6)
		assert.Equal(t, "Unknown", label.LabelName)
		assert.Equal(t, "unknown", label.LabelSlug)
		assert.Equal(t, -6, label.LabelPriority)
	})
}

func TestLabel_SetName(t *testing.T) {
	t.Run("set name", func(t *testing.T) {
		entity := LabelFixtures["landscape"]

		assert.Equal(t, "Landscape", entity.LabelName)
		assert.Equal(t, "landscape", entity.LabelSlug)
		assert.Equal(t, "landscape", entity.CustomSlug)

		entity.SetName("Landschaft")

		assert.Equal(t, "Landschaft", entity.LabelName)
		assert.Equal(t, "landscape", entity.LabelSlug)
		assert.Equal(t, "landschaft", entity.CustomSlug)
	})

	t.Run("new name empty", func(t *testing.T) {
		entity := LabelFixtures["flower"]

		assert.Equal(t, "Flower", entity.LabelName)
		assert.Equal(t, "flower", entity.LabelSlug)
		assert.Equal(t, "flower", entity.CustomSlug)

		entity.SetName("")

		assert.Equal(t, "Flower", entity.LabelName)
		assert.Equal(t, "flower", entity.LabelSlug)
		assert.Equal(t, "flower", entity.CustomSlug)
	})
}

func TestFirstOrCreateLabel(t *testing.T) {
	label := LabelFixtures.Get("flower")
	result := FirstOrCreateLabel(&label)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.LabelName != label.LabelName {
		t.Errorf("LabelName should be the same: %s %s", result.LabelName, label.LabelName)
	}

	if result.LabelSlug != label.LabelSlug {
		t.Errorf("LabelName should be the same: %s %s", result.LabelSlug, label.LabelSlug)
	}
}

func TestLabel_Update(t *testing.T) {
	t.Run("update priority and label slug", func(t *testing.T) {
		classifyLabel := &classify.Label{Name: "classify", Uncertainty: 30, Source: "manual", Priority: 5}
		Label := &Label{LabelName: "label", LabelSlug: "", CustomSlug: "customslug", LabelPriority: 4}

		assert.Equal(t, 4, Label.LabelPriority)
		assert.Equal(t, "", Label.LabelSlug)
		assert.Equal(t, "customslug", Label.CustomSlug)
		assert.Equal(t, "label", Label.LabelName)

		err := Label.UpdateClassify(*classifyLabel)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 5, Label.LabelPriority)
		assert.Equal(t, "customslug", Label.LabelSlug)
		assert.Equal(t, "classify", Label.CustomSlug)
		assert.Equal(t, "Classify", Label.LabelName)
	})
	t.Run("update custom slug", func(t *testing.T) {
		classifyLabel := &classify.Label{Name: "classify", Uncertainty: 30, Source: "manual", Priority: 5}
		Label := &Label{LabelName: "label12", LabelSlug: "labelslug", CustomSlug: "", LabelPriority: 5}

		assert.Equal(t, 5, Label.LabelPriority)
		assert.Equal(t, "labelslug", Label.LabelSlug)
		assert.Equal(t, "", Label.CustomSlug)
		assert.Equal(t, "label12", Label.LabelName)

		err := Label.UpdateClassify(*classifyLabel)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 5, Label.LabelPriority)
		assert.Equal(t, "labelslug", Label.LabelSlug)
		assert.Equal(t, "classify", Label.CustomSlug)
		assert.Equal(t, "Classify", Label.LabelName)

	})
	t.Run("update name and Categories", func(t *testing.T) {
		classifyLabel := &classify.Label{Name: "classify", Uncertainty: 30, Source: "manual", Priority: 5, Categories: []string{"flower", "plant"}}
		Label := &Label{LabelName: "label34", LabelSlug: "labelslug2", CustomSlug: "labelslug2", LabelPriority: 5, LabelCategories: []*Label{LabelFixtures.Pointer("flower")}}

		assert.Equal(t, 5, Label.LabelPriority)
		assert.Equal(t, "labelslug2", Label.LabelSlug)
		assert.Equal(t, "labelslug2", Label.CustomSlug)
		assert.Equal(t, "label34", Label.LabelName)

		err := Label.UpdateClassify(*classifyLabel)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 5, Label.LabelPriority)
		assert.Equal(t, "labelslug2", Label.LabelSlug)
		assert.Equal(t, "classify", Label.CustomSlug)
		assert.Equal(t, "Classify", Label.LabelName)

	})
}
