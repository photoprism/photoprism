package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeywordMap_Get(t *testing.T) {
	t.Run("GetExistingKeyword", func(t *testing.T) {
		r := KeywordFixtures.Get("bridge")
		assert.Equal(t, uint(1000000), r.ID)
		assert.Equal(t, "bridge", r.Keyword)
		assert.IsType(t, Keyword{}, r)
	})
	t.Run("GetNotExistingKeyword", func(t *testing.T) {
		r := KeywordFixtures.Get("Fusion")
		assert.Equal(t, "fusion", r.Keyword)
		assert.IsType(t, Keyword{}, r)
	})
}

func TestKeywordMap_Pointer(t *testing.T) {
	t.Run("GetExistingKeywordPointer", func(t *testing.T) {
		r := KeywordFixtures.Pointer("bridge")
		assert.Equal(t, uint(1000000), r.ID)
		assert.Equal(t, "bridge", r.Keyword)
		assert.IsType(t, &Keyword{}, r)
	})
	t.Run("GetNotExistingKeywordPointer", func(t *testing.T) {
		r := KeywordFixtures.Pointer("sweets")
		assert.Equal(t, "sweets", r.Keyword)
		assert.IsType(t, &Keyword{}, r)
	})
}
