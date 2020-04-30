package entity

import (
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
	entity := LabelFixtures["landscape"]

	assert.Equal(t, "Landscape", entity.LabelName)
	assert.Equal(t, "landscape", entity.LabelSlug)
	assert.Equal(t, "landscape", entity.CustomSlug)

	entity.SetName("Landschaft")

	assert.Equal(t, "Landschaft", entity.LabelName)
	assert.Equal(t, "landscape", entity.LabelSlug)
	assert.Equal(t, "landschaft", entity.CustomSlug)
}
