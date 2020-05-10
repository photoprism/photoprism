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

func TestPhotoKeywords_FirstOrCreate(t *testing.T) {
	m := PhotoKeywordFixtures["1"]
	r := m.FirstOrCreate()
	assert.Equal(t, uint(0xf4244), r.PhotoID)
}
