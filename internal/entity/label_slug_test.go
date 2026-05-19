package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAcceptLabelSlugMatch(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, acceptLabelSlugMatch(nil, "Cat"))
	})
	t.Run("EmptyQuery", func(t *testing.T) {
		candidate := &Label{LabelName: "Cat", LabelSlug: "cat", CustomSlug: "cat"}
		assert.False(t, acceptLabelSlugMatch(candidate, ""))
	})
	t.Run("SameName", func(t *testing.T) {
		candidate := &Label{LabelName: "Cat", LabelSlug: "cat", CustomSlug: "cat"}
		assert.True(t, acceptLabelSlugMatch(candidate, "cat"))
	})
	t.Run("RenamedAwayFromQueriedName", func(t *testing.T) {
		candidate := &Label{LabelName: "Katze", LabelSlug: "cat", CustomSlug: "katze"}
		assert.True(t, acceptLabelSlugMatch(candidate, "Cat"))
	})
	t.Run("HomophoneNotRenamed", func(t *testing.T) {
		// First-created homophone has LabelSlug == CustomSlug; a query for
		// a different name that slugifies the same must not be accepted.
		candidate := &Label{LabelName: "问", LabelSlug: "wen", CustomSlug: "wen"}
		assert.False(t, acceptLabelSlugMatch(candidate, "吻"))
	})
	t.Run("DifferentNameAndDifferentSlug", func(t *testing.T) {
		candidate := &Label{LabelName: "Other", LabelSlug: "other", CustomSlug: "other"}
		assert.False(t, acceptLabelSlugMatch(candidate, "Cat"))
	})
}
