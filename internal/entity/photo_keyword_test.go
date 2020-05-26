package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPhotoKeyword(t *testing.T) {
	t.Run("new keyword", func(t *testing.T) {
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
		t.Fatal("result should not be nil")
	}

	if result.PhotoID != model.PhotoID {
		t.Errorf("PhotoID should be the same: %d %d", result.PhotoID, model.PhotoID)
	}

	if result.KeywordID != model.KeywordID {
		t.Errorf("KeywordID should be the same: %d %d", result.KeywordID, model.KeywordID)
	}
}
