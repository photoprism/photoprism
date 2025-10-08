package thumb

import (
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
)

func TestParseFilter(t *testing.T) {
	t.Run("Vips", func(t *testing.T) {
		assert.Equal(t, ResampleAuto, ParseFilter("", LibVips))
		assert.Equal(t, ResampleAuto, ParseFilter("auto", LibVips))
		assert.Equal(t, ResampleAuto, ParseFilter("blackman", LibVips))
		assert.Equal(t, ResampleAuto, ParseFilter("lanczos", LibVips))
		assert.Equal(t, ResampleAuto, ParseFilter("cubic", LibVips))
		assert.Equal(t, ResampleAuto, ParseFilter("linear", LibVips))
		assert.Equal(t, ResampleAuto, ParseFilter("invalid", LibVips))
	})
	t.Run("Imaging", func(t *testing.T) {
		assert.Equal(t, ResampleLanczos, ParseFilter("", LibImaging))
		assert.Equal(t, ResampleLanczos, ParseFilter("auto", LibImaging))
		assert.Equal(t, ResampleBlackman, ParseFilter("blackman", LibImaging))
		assert.Equal(t, ResampleLanczos, ParseFilter("lanczos", LibImaging))
		assert.Equal(t, ResampleCubic, ParseFilter("cubic", LibImaging))
		assert.Equal(t, ResampleLinear, ParseFilter("linear", LibImaging))
		assert.Equal(t, ResampleLanczos, ParseFilter("invalid", LibImaging))
	})
}

func TestResampleFilter_StringAndVips(t *testing.T) {
	// String
	assert.Equal(t, "blackman", ResampleBlackman.String())
	assert.Equal(t, "lanczos", ResampleLanczos.String())
	assert.Equal(t, "cubic", ResampleCubic.String())
	assert.Equal(t, "linear", ResampleLinear.String())
	assert.Equal(t, "nearest", ResampleNearest.String())

	// Vips mapping
	assert.Equal(t, vips.KernelLanczos3, ResampleBlackman.Vips())
	assert.Equal(t, vips.KernelLanczos3, ResampleLanczos.Vips())
	assert.Equal(t, vips.KernelLanczos3, ResampleAuto.Vips())
	assert.Equal(t, vips.KernelCubic, ResampleCubic.Vips())
	assert.Equal(t, vips.KernelLinear, ResampleLinear.Vips())
	assert.Equal(t, vips.KernelNearest, ResampleNearest.Vips())
}
