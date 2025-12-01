package entity

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindKeyword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		FlushKeywordCache()

		keywordName := fmt.Sprintf("keyword-cache-%d", time.Now().UnixNano())
		keyword := NewKeyword(keywordName)

		require.NoError(t, keyword.Save())

		uncached, err := FindKeyword(keyword.Keyword, false)
		assert.NoError(t, err)
		assert.Equal(t, keyword.Keyword, uncached.Keyword)
		assert.Equal(t, keyword.ID, uncached.ID)

		cached, err := FindKeyword(keyword.Keyword, true)
		assert.NoError(t, err)
		assert.Equal(t, keyword.Keyword, cached.Keyword)
		assert.Equal(t, keyword.ID, cached.ID)
	})
	t.Run("NotFound", func(t *testing.T) {
		FlushKeywordCache()

		missingName := fmt.Sprintf("missing-keyword-%d", time.Now().UnixNano())
		result, err := FindKeyword(missingName, true)
		assert.Error(t, err)
		assert.NotNil(t, result)

		result, err = FindKeyword(missingName, false)
		assert.Error(t, err)
		assert.NotNil(t, result)
	})
	t.Run("EmptyName", func(t *testing.T) {
		FlushKeywordCache()

		result, err := FindKeyword("", true)
		assert.Error(t, err)
		assert.NotNil(t, result)
	})
}

func TestFindPhotoKeyword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		FlushPhotoKeywordCache()
		require.NoError(t, CachePhotoKeywords())

		fixture := PhotoKeywordFixtures["3"]
		cached, err := FindPhotoKeyword(fixture.PhotoID, fixture.KeywordID, true)
		assert.NoError(t, err)
		assert.Equal(t, fixture.KeywordID, cached.KeywordID)
		assert.Equal(t, fixture.PhotoID, cached.PhotoID)

		FlushPhotoKeywordCache()
		cached, err = FindPhotoKeyword(fixture.PhotoID, fixture.KeywordID, true)
		assert.NoError(t, err)
		assert.Equal(t, fixture.KeywordID, cached.KeywordID)
		assert.Equal(t, fixture.PhotoID, cached.PhotoID)
	})
	t.Run("NotFound", func(t *testing.T) {
		FlushPhotoKeywordCache()
		missingPhotoID := uint(5000000)
		missingKeywordID := uint(6000000)

		result, err := FindPhotoKeyword(missingPhotoID, missingKeywordID, true)
		assert.Error(t, err)
		assert.NotNil(t, result)

		result, err = FindPhotoKeyword(missingPhotoID, missingKeywordID, false)
		assert.Error(t, err)
		assert.NotNil(t, result)
	})
	t.Run("InvalidID", func(t *testing.T) {
		FlushPhotoKeywordCache()
		result, err := FindPhotoKeyword(0, 0, true)
		assert.Error(t, err)
		assert.NotNil(t, result)

		result, err = FindPhotoKeyword(0, 0, false)
		assert.Error(t, err)
		assert.NotNil(t, result)
	})
}

func TestCachePhotoKeywords(t *testing.T) {
	FlushPhotoKeywordCache()
	require.NoError(t, CachePhotoKeywords())

	fixture := PhotoKeywordFixtures["3"]
	_, found := photoKeywordCache.Get(photoKeywordCacheKey(fixture.PhotoID, fixture.KeywordID))
	assert.True(t, found)
}

func TestFlushCachedPhotoKeyword(t *testing.T) {
	FlushPhotoKeywordCache()
	fixture := PhotoKeywordFixtures["3"]
	cacheKey := photoKeywordCacheKey(fixture.PhotoID, fixture.KeywordID)
	photoKeywordCache.SetDefault(cacheKey, fixture)

	FlushCachedPhotoKeyword(&fixture)

	_, found := photoKeywordCache.Get(cacheKey)
	assert.False(t, found)

	// Ensure nil inputs are safe.
	FlushCachedPhotoKeyword(nil)
}
