package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexOptionsNone(t *testing.T) {
	result := IndexOptionsNone()
	assert.Equal(t, false, result.Rescan)
	assert.Equal(t, false, result.Convert)
}

func TestIndexOptions_SkipUnchanged(t *testing.T) {
	result := IndexOptionsNone()
	assert.True(t, result.SkipUnchanged())
	result.Rescan = true
	assert.False(t, result.SkipUnchanged())
}
