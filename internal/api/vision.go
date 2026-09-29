package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// abortVisionApiDisabled answers with a Vision API error and returns true if the Vision API is disabled.
func abortVisionApiDisabled(c *gin.Context) bool {
	if get.Config().VisionApi() {
		return false
	}

	c.AbortWithStatusJSON(http.StatusForbidden, vision.NewApiError(rnd.UUID(), http.StatusForbidden))

	return true
}
