package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileTypeOffset(t *testing.T) {
	t.Run("Mp4vAvc1Mp4", func(t *testing.T) {
		index, err := FileTypeOffset("testdata/mp4v-avc1.mp4", CompatibleBrands)
		require.NoError(t, err)
		assert.Equal(t, 0, index)
	})
	t.Run("IsomAvc1Mp4", func(t *testing.T) {
		index, err := FileTypeOffset("testdata/isom-avc1.mp4", CompatibleBrands)
		require.NoError(t, err)
		assert.Equal(t, 0, index)
	})
	t.Run("ImageIsomAvc1Jpg", func(t *testing.T) {
		index, err := FileTypeOffset("testdata/image-isom-avc1.jpg", CompatibleBrands)
		require.NoError(t, err)
		assert.Equal(t, 23209, index)
	})
}
