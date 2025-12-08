package vision

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/thumb"
)

func TestResolution(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		result := Resolution("invalid")
		assert.Equal(t, DefaultResolution, result)
	})
	t.Run("Facenet", func(t *testing.T) {
		result := Resolution(ModelTypeFace)
		assert.Equal(t, FacenetModel.Resolution, result)
	})
	t.Run("Nasnet", func(t *testing.T) {
		result := Resolution(ModelTypeLabels)
		assert.Equal(t, 224, result)
	})
}

func TestThumb(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		size := Thumb("invalid")
		assert.Equal(t, thumb.SizeTile224, size)
	})
	t.Run("Facenet", func(t *testing.T) {
		size := Thumb(ModelTypeFace)
		assert.Equal(t, thumb.SizeTile224, size)
	})
	t.Run("Nasnet", func(t *testing.T) {
		size := Thumb(ModelTypeLabels)
		assert.Equal(t, thumb.SizeTile224, size)
	})
	t.Run("Caption", func(t *testing.T) {
		size := Thumb(ModelTypeCaption)
		assert.Equal(t, thumb.SizeFit720, size)
	})
}
