package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViewerThumbs(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		result := ViewerThumbs(2000, 1500, "011df944f313a05f89d170a561fad09ce6cef44e", "https://example.com", "12345678")

		assert.Equal(t, 720, result.Fit720.W)
		assert.Equal(t, 540, result.Fit720.H)
		assert.Equal(t, "https://example.com/t/011df944f313a05f89d170a561fad09ce6cef44e/12345678/fit_720", result.Fit720.Src)
		assert.Equal(t, 1280, result.Fit1280.W)
		assert.Equal(t, 960, result.Fit1280.H)
		assert.Equal(t, "https://example.com/t/011df944f313a05f89d170a561fad09ce6cef44e/12345678/fit_1280", result.Fit1280.Src)
		assert.Equal(t, 1600, result.Fit1920.W)
		assert.Equal(t, 1200, result.Fit1920.H)
		assert.Equal(t, "https://example.com/t/011df944f313a05f89d170a561fad09ce6cef44e/12345678/fit_1920", result.Fit1920.Src)
		assert.Equal(t, 2000, result.Fit2560.W)
		assert.Equal(t, 1500, result.Fit2560.H)
		assert.Equal(t, "https://example.com/t/011df944f313a05f89d170a561fad09ce6cef44e/12345678/fit_2560", result.Fit2560.Src)
		assert.Equal(t, 2000, result.Fit4096.W)
		assert.Equal(t, 1500, result.Fit4096.H)
		assert.Equal(t, "https://example.com/t/011df944f313a05f89d170a561fad09ce6cef44e/12345678/fit_2560", result.Fit2560.Src)
		assert.Equal(t, 2000, result.Fit5120.W)
		assert.Equal(t, 1500, result.Fit5120.H)
		assert.Equal(t, "https://example.com/t/011df944f313a05f89d170a561fad09ce6cef44e/12345678/fit_2560", result.Fit5120.Src)
		assert.Equal(t, 2000, result.Fit7680.W)
		assert.Equal(t, 1500, result.Fit7680.H)
		assert.Equal(t, "https://example.com/t/011df944f313a05f89d170a561fad09ce6cef44e/12345678/fit_2560", result.Fit7680.Src)
	})
}
