package viewer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloadUrl(t *testing.T) {
	apiUri := "/api/v1"

	t.Run("WithToken", func(t *testing.T) {
		dlToken := "3tcsggxy"
		result := DownloadUrl("d2b4a5d18276f96f1b5a1bf17fd82d6fab3807f2", apiUri, dlToken)
		assert.Equal(t, "/api/v1/dl/d2b4a5d18276f96f1b5a1bf17fd82d6fab3807f2?t=3tcsggxy", result)
	})

	t.Run("NoToken", func(t *testing.T) {
		dlToken := ""
		result := DownloadUrl("653cd9e5754e98d899e9ba30c9075da4ebb16141", apiUri, dlToken)
		assert.Equal(t, "/api/v1/dl/653cd9e5754e98d899e9ba30c9075da4ebb16141?t=", result)
	})
}
