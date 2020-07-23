package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_ExceedsLimit(t *testing.T) {
	Size = 1024
	SizeUncached = 2048

	fit4096 := Types["fit_4096"]
	assert.True(t, fit4096.ExceedsSizeUncached())

	fit2048 := Types["fit_2048"]
	assert.False(t, fit2048.ExceedsSizeUncached())

	tile500 := Types["tile_500"]
	assert.False(t, tile500.ExceedsSizeUncached())

	Size = 2048
	SizeUncached = 7680
}

func TestType_ExceedsSize(t *testing.T) {
	Size = 1024
	SizeUncached = 2048

	fit4096 := Types["fit_4096"]
	assert.True(t, fit4096.ExceedsSize())

	fit2048 := Types["fit_2048"]
	assert.True(t, fit2048.ExceedsSize())

	tile500 := Types["tile_500"]
	assert.False(t, tile500.ExceedsSize())

	Size = 2048
	SizeUncached = 7680
}

func TestType_SkipPreRender(t *testing.T) {
	Size = 1024
	SizeUncached = 2048

	fit4096 := Types["fit_4096"]
	assert.True(t, fit4096.OnDemand())

	fit2048 := Types["fit_2048"]
	assert.True(t, fit2048.OnDemand())

	tile500 := Types["tile_500"]
	assert.False(t, tile500.OnDemand())

	Size = 2048
	SizeUncached = 7680
}

func TestResampleFilter_Imaging(t *testing.T) {
	t.Run("Blackman", func(t *testing.T) {
		r := ResampleBlackman.Imaging()
		assert.Equal(t, float64(3), r.Support)
	})
	t.Run("Cubic", func(t *testing.T) {
		r := ResampleCubic.Imaging()
		assert.Equal(t, float64(2), r.Support)
	})
	t.Run("Linear", func(t *testing.T) {
		r := ResampleLinear.Imaging()
		assert.Equal(t, float64(1), r.Support)
	})
}

func TestFinde(t *testing.T) {
	t.Run("2048", func(t *testing.T) {
		tName, tType := Find(2048)
		assert.Equal(t, "fit_2048", tName)
		assert.Equal(t, 2048, tType.Width)
		assert.Equal(t, 2048, tType.Height)
	})

	t.Run("2000", func(t *testing.T) {
		tName, tType := Find(2000)
		assert.Equal(t, "fit_1920", tName)
		assert.Equal(t, 1920, tType.Width)
		assert.Equal(t, 1200, tType.Height)
	})
}
