package context

import (
	"testing"

	"github.com/photoprism/photoprism/internal/fsutil"
	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	ctx := CliTestContext()

	assert.True(t, ctx.IsSet("assets-path"))
	assert.False(t, ctx.Bool("debug"))

	c := NewContext(ctx)

	assert.IsType(t, new(Context), c)

	assert.Equal(t, fsutil.ExpandedFilename("../../assets"), c.AssetsPath())
	assert.False(t, c.Debug())
}
