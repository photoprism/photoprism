package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_ExceedsLimit(t *testing.T) {
	Size = 1024
	Limit = 2048

	fit3840 := Types["fit_3840"]
	assert.True(t, fit3840.ExceedsLimit())

	fit2048 := Types["fit_2048"]
	assert.False(t, fit2048.ExceedsLimit())

	tile500 := Types["tile_500"]
	assert.False(t, tile500.ExceedsLimit())

	Size = 3840
	Limit = 3840
}

func TestType_SkipPreRender(t *testing.T) {
	Size = 1024
	Limit = 2048

	fit3840 := Types["fit_3840"]
	assert.True(t, fit3840.OnDemand())

	fit2048 := Types["fit_2048"]
	assert.True(t, fit2048.OnDemand())

	tile500 := Types["tile_500"]
	assert.False(t, tile500.OnDemand())

	Size = 3840
	Limit = 3840
}
