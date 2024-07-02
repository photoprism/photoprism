package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SearchFaces finds and returns faces as JSON.
//
// GET /api/v1/faces
func SearchFaces(router *gin.RouterGroup) {
	router.GET("/faces", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePeople, acl.ActionSearch)

		// Abort if permission was not granted.
		if s.Abort(c) {
			return
		}

		var f form.SearchFaces

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		result, err := search.Faces(f)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		AddCountHeader(c, len(result))
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c, s)

		c.JSON(http.StatusOK, result)
	})
}
