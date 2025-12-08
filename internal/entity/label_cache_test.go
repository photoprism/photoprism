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

func TestFindPhotoLabel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if err := CachePhotoLabels(); err != nil {
			t.Fatal(err)
		}

		// See PhotoFixtures and LabelFixtures for test data.
		m := &PhotoLabel{PhotoID: 1000000, LabelID: 1000001}

		cached, err := FindPhotoLabel(m.PhotoID, m.LabelID, true)

		assert.NoError(t, err)
		assert.Equal(t, m.LabelID, cached.LabelID)
		assert.Equal(t, m.PhotoID, cached.PhotoID)
		assert.Equal(t, SrcImage, cached.LabelSrc)
		assert.Equal(t, 38, cached.Uncertainty)

		FlushPhotoLabelCache()

		cached, err = FindPhotoLabel(m.PhotoID, m.LabelID, true)

		assert.NoError(t, err)
		assert.Equal(t, m.LabelID, cached.LabelID)
		assert.Equal(t, m.PhotoID, cached.PhotoID)
		assert.Equal(t, SrcImage, cached.LabelSrc)
		assert.Equal(t, 38, cached.Uncertainty)
	})
	t.Run("NotFound", func(t *testing.T) {
		result, err := FindPhotoLabel(1, 99999999, true)
		assert.Error(t, err)
		assert.NotNil(t, result)
		result, err = FindPhotoLabel(1, 99999999, false)
		assert.Error(t, err)
		assert.NotNil(t, result)
		result, err = FindPhotoLabel(1, 99999999, true)
		assert.Error(t, err)
		assert.NotNil(t, result)
	})
	t.Run("InvalidID", func(t *testing.T) {
		result, err := FindPhotoLabel(0, 0, true)
		assert.Error(t, err)
		assert.NotNil(t, result)
		result, err = FindPhotoLabel(0, 0, false)
		assert.Error(t, err)
		assert.NotNil(t, result)
		result, err = FindPhotoLabel(0, 0, true)
		assert.Error(t, err)
		assert.NotNil(t, result)
	})
}
