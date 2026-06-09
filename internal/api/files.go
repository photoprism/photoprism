package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
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

		// Abort if permission is not granted.
		if s.Abort(c) {
			return
		}

		hash := clean.Token(c.Param("hash"))

		// Limit results to files within the session's shared scope, consistent with how photo
		// search filters results. Files outside the scope are reported as not found.
		if visible, err := search.FileVisibleToSession(hash, s); err != nil || !visible {
			AbortEntityNotFound(c)
			return
		}

		p, err := query.FileByHash(hash)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, p)
	})
}
