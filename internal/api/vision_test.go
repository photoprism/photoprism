package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestAbortVisionApiDisabled checks that the helper answers and aborts only when the Vision API is disabled.
func TestAbortVisionApiDisabled(t *testing.T) {
	conf := get.Config()
	orig := conf.Options().VisionApi
	t.Cleanup(func() { conf.Options().VisionApi = orig })

	newContext := func() (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/vision/labels", nil)
		return c, w
	}

	t.Run("Enabled", func(t *testing.T) {
		conf.Options().VisionApi = true
		c, w := newContext()
		assert.False(t, abortVisionApiDisabled(c))
		assert.False(t, c.IsAborted())
		assert.Zero(t, w.Body.Len())
	})
	t.Run("Disabled", func(t *testing.T) {
		conf.Options().VisionApi = false
		c, w := newContext()
		assert.True(t, abortVisionApiDisabled(c))
		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var resp vision.ApiResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.True(t, rnd.IsUUID(resp.Id), resp.Id)
		assert.Equal(t, http.StatusForbidden, resp.Code)
		assert.Equal(t, "Forbidden", resp.Error)
	})
}
