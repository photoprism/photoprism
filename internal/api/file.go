package api

import (
	"net/http"

	"github.com/photoprism/photoprism/pkg/clean"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/query"
)

// GetFile returns file details as JSON.
//
// Route: GET /api/v1/files/:hash
// Params:
// - hash (string) SHA-1 hash of the file
func GetFile(router *gin.RouterGroup) {
	router.GET("/files/:hash", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceFiles, acl.ActionRead)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		p, err := query.FileByHash(clean.Token(c.Param("hash")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, p)
	})
}
