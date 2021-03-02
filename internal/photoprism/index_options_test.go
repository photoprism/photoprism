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

func TestIndexOptionsSingle(t *testing.T) {
	r := IndexOptionsSingle()
	assert.Equal(t, false, r.Stack)
	assert.Equal(t, true, r.Convert)
	assert.Equal(t, true, r.Rescan)
}
