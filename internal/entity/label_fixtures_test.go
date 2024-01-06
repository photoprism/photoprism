package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelMap_Get(t *testing.T) {
	t.Run("get existing label", func(t *testing.T) {
		r := LabelFixtures.Get("landscape")
		assert.Equal(t, "ls6sg6b1wowuy3c2", r.LabelUID)
		assert.Equal(t, "landscape", r.LabelSlug)
		assert.IsType(t, Label{}, r)
	})
	t.Run("get not existing label", func(t *testing.T) {
		r := LabelFixtures.Get("monstera")
		assert.Equal(t, "monstera", r.LabelSlug)
		assert.IsType(t, Label{}, r)
	})
}

func TestLabelMap_Pointer(t *testing.T) {
	t.Run("get existing label pointer", func(t *testing.T) {
		r := LabelFixtures.Pointer("landscape")
		assert.Equal(t, "ls6sg6b1wowuy3c2", r.LabelUID)
		assert.Equal(t, "landscape", r.LabelSlug)
		assert.IsType(t, &Label{}, r)
	})
	t.Run("get not existing label pointer", func(t *testing.T) {
		r := LabelFixtures.Pointer("monstera Leaf")
		assert.Equal(t, "monstera-leaf", r.LabelSlug)
		assert.IsType(t, &Label{}, r)
	})
}

func TestLabelMap_PhotoLabel(t *testing.T) {
	r := LabelFixtures.PhotoLabel(123, "landscape", 25, "")
	assert.IsType(t, PhotoLabel{}, r)
}
