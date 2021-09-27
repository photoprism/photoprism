package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexOptionsNone(t *testing.T) {
	opt := IndexOptionsNone()

	assert.Equal(t, "", opt.Path)
	assert.Equal(t, false, opt.Rescan)
	assert.Equal(t, false, opt.Convert)
	assert.Equal(t, false, opt.Stack)
	assert.Equal(t, false, opt.FacesOnly)
}

func TestIndexOptions_SkipUnchanged(t *testing.T) {
	opt := IndexOptionsNone()

	assert.True(t, opt.SkipUnchanged())

	opt.Rescan = true

	assert.False(t, opt.SkipUnchanged())
}

func TestIndexOptionsSingle(t *testing.T) {
	opt := IndexOptionsSingle()

	assert.Equal(t, false, opt.Stack)
	assert.Equal(t, true, opt.Convert)
	assert.Equal(t, true, opt.Rescan)
}

func TestIndexOptionsFacesOnly(t *testing.T) {
	opt := IndexOptionsFacesOnly()

	assert.Equal(t, "/", opt.Path)
	assert.Equal(t, true, opt.Rescan)
	assert.Equal(t, true, opt.Convert)
	assert.Equal(t, true, opt.Stack)
	assert.Equal(t, true, opt.FacesOnly)
}
