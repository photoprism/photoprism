package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
)

// GetFile returns file details as JSON.
//
//	@Summary	returns file details as JSON
//	@Id			GetFile
//	@Tags		Files
//	@Produce	json
//	@Success	200				{object}	entity.File
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		hash			path		string	true	"hash (string) SHA-1 hash of the file"
//	@Router		/api/v1/files/{hash} [get]
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
