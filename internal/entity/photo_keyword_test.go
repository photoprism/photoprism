package entity

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPhotoKeyword(t *testing.T) {
	t.Run("NewKeyword", func(t *testing.T) {
		m := NewPhotoKeyword(uint(3), uint(8))
		assert.Equal(t, uint(3), m.PhotoID)
		assert.Equal(t, uint(8), m.KeywordID)
	})
}

func TestPhotoKeyword_TableName(t *testing.T) {
	photoKeyword := &PhotoKeyword{}
	tableName := photoKeyword.TableName()

	assert.Equal(t, "photos_keywords", tableName)
}

func TestFirstOrCreatePhotoKeyword(t *testing.T) {
	model := PhotoKeywordFixtures["1"]
	result := FirstOrCreatePhotoKeyword(&model)

	if result == nil {
		t.Fatal("result must not be nil")
	}

	if result.PhotoID != model.PhotoID {
		t.Errorf("PhotoID should be the same: %d %d", result.PhotoID, model.PhotoID)
	}

	if result.KeywordID != model.KeywordID {
		t.Errorf("KeywordID should be the same: %d %d", result.KeywordID, model.KeywordID)
	}
}

func TestPhotoKeyword_Delete(t *testing.T) {
	FlushPhotoKeywordCache()
	photo := &Photo{}
	require.NoError(t, Db().First(photo).Error)
	keyword := NewKeyword(fmt.Sprintf("photo-keyword-delete-%d", time.Now().UnixNano()))
	require.NoError(t, keyword.Save())

	relation := NewPhotoKeyword(photo.ID, keyword.ID)
	require.NoError(t, relation.Create())

	photoKeywordCache.SetDefault(relation.CacheKey(), *relation)

	require.NoError(t, relation.Delete())

	_, found := photoKeywordCache.Get(relation.CacheKey())
	assert.False(t, found)
}
