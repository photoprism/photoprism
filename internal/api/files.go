package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
)

// GetFile returns file details as JSON.
//
// The request parameters are:
//
//   - hash (string) SHA-1 hash of the file
//
// GET /api/v1/files/:hash
func GetFile(router *gin.RouterGroup) {
	router.GET("/files/:hash", func(c *gin.Context) {
		s := Auth(c, acl.ResourceFiles, acl.ActionView)

		// Abort if permission was not granted.
		if s.Abort(c) {
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
