package photoprism

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIndexOptionsNone(t *testing.T) {
	result := IndexOptionsNone()
	assert.Equal(t, false, result.UpdateCamera)
	assert.Equal(t, false, result.UpdateDate)
	assert.Equal(t, false, result.UpdateColors)
}

func TestIndexOptions_UpdateAny(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		result := IndexOptionsAll()
		assert.True(t, result.UpdateAny())
	})

	t.Run("true", func(t *testing.T) {
		result := IndexOptionsNone()
		assert.False(t, result.UpdateAny())
	})
}

func TestIndexOptions_SkipUnchanged(t *testing.T) {
	result := IndexOptionsNone()
	assert.True(t, result.SkipUnchanged())
}
