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
	t.Run("Homophone", func(t *testing.T) {
		label1 := FirstOrCreateLabel(NewLabel("老板", 10))
		if !assert.NotNil(t, label1) {
			t.Fatal("label should not be nil")
		}
		label2 := FirstOrCreateLabel(NewLabel("老伴", 10))
		if !assert.NotNil(t, label2) {
			t.Fatal("label should not be nil")
		}
		label3 := FirstOrCreateLabel(NewLabel("lao-ban", 10))
		if !assert.NotNil(t, label3) {
			t.Fatal("label should not be nil")
		}

		uncached1, findErr := FindLabel("老板", false)
		assert.NoError(t, findErr)
		assert.Equal(t, "老板", uncached1.LabelName)
		uncached2, findErr := FindLabel("老伴", false)
		assert.NoError(t, findErr)
		assert.Equal(t, "老伴", uncached2.LabelName)
		uncached3, findErr := FindLabel("lao-ban", false)
		assert.NoError(t, findErr)
		assert.Equal(t, "Lao-Ban", uncached3.LabelName)

		cached1, cacheErr := FindLabel("老板", true)

		assert.NoError(t, cacheErr)
		assert.Equal(t, "老板", cached1.LabelName)
		assert.Equal(t, uncached1.LabelSlug, cached1.LabelSlug)
		assert.Equal(t, uncached1.ID, cached1.ID)
		assert.Equal(t, uncached1.LabelUID, cached1.LabelUID)

		cached2, cacheErr := FindLabel("老伴", true)

		assert.NoError(t, cacheErr)
		assert.Equal(t, "老伴", cached2.LabelName)
		assert.Equal(t, uncached2.LabelSlug, cached2.LabelSlug)
		assert.Equal(t, uncached2.ID, cached2.ID)
		assert.Equal(t, uncached2.LabelUID, cached2.LabelUID)

		cached3, cacheErr := FindLabel("lao-ban", true)

		assert.NoError(t, cacheErr)
		assert.Equal(t, "Lao-Ban", cached3.LabelName)
		assert.Equal(t, uncached3.LabelSlug, cached3.LabelSlug)
		assert.Equal(t, uncached3.ID, cached3.ID)
		assert.Equal(t, uncached3.LabelUID, cached3.LabelUID)

		assert.NoError(t, UnscopedDb().Delete(&label1).Error)
		assert.NoError(t, UnscopedDb().Delete(&label2).Error)
		assert.NoError(t, UnscopedDb().Delete(&label3).Error)
	})

	t.Run("HomophoneCacheInvalidation", func(t *testing.T) {
		label1 := FirstOrCreateLabel(NewLabel("老板", 10))
		if !assert.NotNil(t, label1) {
			t.Fatal("label should not be nil")
		}
		// Cache it before the MaxHomophone is populated
		uncached1, findErr := FindLabel("老板", false)
		assert.NoError(t, findErr)
		assert.Equal(t, "老板", uncached1.LabelName)

		label2 := FirstOrCreateLabel(NewLabel("老伴", 10))
		if !assert.NotNil(t, label2) {
			t.Fatal("label should not be nil")
		}
		label3 := FirstOrCreateLabel(NewLabel("lao-ban", 10))
		if !assert.NotNil(t, label3) {
			t.Fatal("label should not be nil")
		}

		// if the cache hasn't been invalidated, then this will fail.
		uncached2, findErr := FindLabel("老伴", false)
		assert.NoError(t, findErr)
		assert.Equal(t, "老伴", uncached2.LabelName)

		uncached3, findErr := FindLabel("lao-ban", false)
		assert.NoError(t, findErr)
		assert.Equal(t, "Lao-Ban", uncached3.LabelName)

		cached1, cacheErr := FindLabel("老板", true)

		assert.NoError(t, cacheErr)
		assert.Equal(t, "老板", cached1.LabelName)
		assert.Equal(t, uncached1.LabelSlug, cached1.LabelSlug)
		assert.Equal(t, uncached1.ID, cached1.ID)
		assert.Equal(t, uncached1.LabelUID, cached1.LabelUID)

		cached2, cacheErr := FindLabel("老伴", true)

		assert.NoError(t, cacheErr)
		assert.Equal(t, "老伴", cached2.LabelName)
		assert.Equal(t, uncached2.LabelSlug, cached2.LabelSlug)
		assert.Equal(t, uncached2.ID, cached2.ID)
		assert.Equal(t, uncached2.LabelUID, cached2.LabelUID)

		cached3, cacheErr := FindLabel("lao-ban", true)

		assert.NoError(t, cacheErr)
		assert.Equal(t, "Lao-Ban", cached3.LabelName)
		assert.Equal(t, uncached3.LabelSlug, cached3.LabelSlug)
		assert.Equal(t, uncached3.ID, cached3.ID)
		assert.Equal(t, uncached3.LabelUID, cached3.LabelUID)

		assert.NoError(t, UnscopedDb().Delete(&label1).Error)
		assert.NoError(t, UnscopedDb().Delete(&label2).Error)
		assert.NoError(t, UnscopedDb().Delete(&label3).Error)
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
