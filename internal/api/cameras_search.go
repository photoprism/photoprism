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

// SearchCameras finds and returns cameras as JSON.
//
//	@Summary	finds and returns cameras as JSON
//	@Id			SearchCameras
//	@Tags		Cameras
//	@Produce	json
//	@Success	200				{array}		search.Camera
//	@Header		200				{number}	X-Count		"The actual number of cameras returned"
//	@Header		200				{number}	X-Limit		"The limit of the number of cameras to be returned"
//	@Header		200				{number}	X-Offset	"The offset that was used"
//	@Failure	401,429,403,400	{object}	i18n.Response
//	@Param		count			query		int		true	"maximum number of results"	minimum(1)	maximum(100000)
//	@Param		offset			query		int		false	"search result offset"		minimum(0)	maximum(100000)
//	@Param		nomake			query		bool	false	"show where make is blank"
//	@Param		q				query		string	false	"search query"
//	@Router		/api/v1/cameras [get]
func SearchCameras(router *gin.RouterGroup) {
	router.GET("/cameras", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCameras, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		var frm form.SearchCameras

		err := c.MustBindWith(&frm, binding.Form)

		if err != nil {
			AbortBadRequest(c, err)
			return
		}

		// Search matching cameras.
		result, err := search.Cameras(frm)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		AddCountHeader(c, len(result))
		AddLimitHeader(c, frm.Count)
		AddOffsetHeader(c, frm.Offset)
		AddTokenHeaders(c, s)

		c.JSON(http.StatusOK, result)
	})
}
