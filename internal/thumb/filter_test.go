package thumb

import (
	"testing"

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
