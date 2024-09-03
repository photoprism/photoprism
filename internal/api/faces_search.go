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
//	@Summary	finds and returns faces as JSON
//	@Id			SearchFaces
//	@Tags		Faces
//	@Produce	json
//	@Success	200					{object}	search.FaceResults
//	@Failure	400,401,403,429,404	{object}	i18n.Response
//	@Param		count				query		int		true	"maximum number of results"	minimum(1)	maximum(100000)
//	@Param		offset				query		int		false	"search result offset"		minimum(0)	maximum(100000)
//	@Param		order				query		string	false	"sort order"				Enums(subject, added, samples)
//	@Param		hidden				query		string	false	"show hidden"				Enums(yes, no)
//	@Param		unknown				query		string	false	"show unknown"				Enums(yes, no)
//	@Router		/api/v1/faces [get]
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
