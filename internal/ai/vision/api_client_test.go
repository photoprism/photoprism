package vision

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/media/http/scheme"
)

func TestNewApiRequest(t *testing.T) {
	t.Run("Data", func(t *testing.T) {
		thumbnails := Files{examplesPath + "/chameleon_lime.jpg"}
		result, err := NewApiRequestImages(thumbnails, scheme.Data)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// t.Logf("request: %#v", result)

		if result != nil {
			json, jsonErr := result.JSON()
			assert.NoError(t, jsonErr)
			assert.NotEmpty(t, json)
			// t.Logf("json: %s", json)
		}
	})
	t.Run("Https", func(t *testing.T) {
		thumbnails := Files{examplesPath + "/chameleon_lime.jpg"}
		result, err := NewApiRequestImages(thumbnails, scheme.Https)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// t.Logf("request: %#v", result)
		if result != nil {
			json, jsonErr := result.JSON()
			assert.NoError(t, jsonErr)
			assert.NotEmpty(t, json)
			t.Logf("json: %s", json)
		}
	})
}
