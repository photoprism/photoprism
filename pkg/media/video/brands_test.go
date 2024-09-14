package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileTypeOffset(t *testing.T) {
	t.Run("mp4v-avc1.mp4", func(t *testing.T) {
		index, err := FileTypeOffset("testdata/mp4v-avc1.mp4", CompatibleBrands)
		require.NoError(t, err)
		assert.Equal(t, 0, index)
	})
	t.Run("isom-avc1.mp4", func(t *testing.T) {
		index, err := FileTypeOffset("testdata/isom-avc1.mp4", CompatibleBrands)
		require.NoError(t, err)
		assert.Equal(t, 0, index)
	})
	t.Run("image-isom-avc1.jpg", func(t *testing.T) {
		index, err := FileTypeOffset("testdata/image-isom-avc1.jpg", CompatibleBrands)
		require.NoError(t, err)
		assert.Equal(t, 23209, index)
	})
}
