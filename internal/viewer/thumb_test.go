package viewer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/thumb"
)

func TestNewThumb(t *testing.T) {
	fileHash := "d2b4a5d18276f96f1b5a1bf17fd82d6fab3807f2"
	contentUri := "/content"
	previewToken := "preview-token"

	t.Run("Fit1280", func(t *testing.T) {
		result := NewThumb(1920, 1080, fileHash, thumb.Sizes[thumb.Fit1280], contentUri, previewToken)

		assert.Equal(t, 1280, result.W)
		assert.Equal(t, 720, result.H)
		assert.Equal(t, "/content/t/d2b4a5d18276f96f1b5a1bf17fd82d6fab3807f2/preview-token/fit_1280", result.Src)
	})

	t.Run("Fit3840", func(t *testing.T) {
		result := NewThumb(1920, 1080, fileHash, thumb.Sizes[thumb.Fit3840], contentUri, previewToken)

		assert.Equal(t, 1920, result.W)
		assert.Equal(t, 1080, result.H)
		assert.Equal(t, "/content/t/d2b4a5d18276f96f1b5a1bf17fd82d6fab3807f2/preview-token/fit_3840", result.Src)
	})
}
