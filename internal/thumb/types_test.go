package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_ExceedsLimit(t *testing.T) {
	PreRenderSize    = 1024
	MaxRenderSize    = 2048

	fit3840 := Types["fit_3840"]
	assert.True(t, fit3840.ExceedsLimit())

	fit2048 := Types["fit_2048"]
	assert.False(t, fit2048.ExceedsLimit())

	tile500 := Types["tile_500"]
	assert.False(t, tile500.ExceedsLimit())

	PreRenderSize    = 3840
	MaxRenderSize    = 3840
}

func TestType_SkipPreRender(t *testing.T) {
	PreRenderSize    = 1024
	MaxRenderSize    = 2048

	fit3840 := Types["fit_3840"]
	assert.True(t, fit3840.SkipPreRender())

	fit2048 := Types["fit_2048"]
	assert.True(t, fit2048.SkipPreRender())

	tile500 := Types["tile_500"]
	assert.False(t, tile500.SkipPreRender())

	PreRenderSize    = 3840
	MaxRenderSize    = 3840
}
