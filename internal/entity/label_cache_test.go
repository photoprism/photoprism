package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindLabel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		label := &Label{LabelSlug: "find-me-label", LabelName: "Find Me"}

		if err := label.Save(); err != nil {
			t.Fatal(err)
		}

		uncached, findErr := FindLabel("find-me-label", false)

		assert.NoError(t, findErr)
		assert.Equal(t, "Find Me", uncached.LabelName)

		cached, cacheErr := FindLabel("find-me-label", true)

		assert.NoError(t, cacheErr)
		assert.Equal(t, "Find Me", cached.LabelName)
		assert.Equal(t, uncached.LabelSlug, cached.LabelSlug)
		assert.Equal(t, uncached.ID, cached.ID)
		assert.Equal(t, uncached.LabelUID, cached.LabelUID)
	})
	t.Run("NotFound", func(t *testing.T) {
		result, err := FindLabel("XXX", true)
		assert.Error(t, err)
		assert.NotNil(t, result)
	})
	t.Run("EmptyName", func(t *testing.T) {
		result, err := FindLabel("", true)
		assert.Error(t, err)
		assert.NotNil(t, result)
	})
}
