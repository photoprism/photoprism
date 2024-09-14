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

// SearchLabels finds and returns labels as JSON.
//
//	@Summary	finds and returns labels as JSON
//	@Id			SearchLabels
//	@Tags		Labels
//	@Produce	json
//	@Success	200				{object}	search.Label
//	@Failure	401,429,403,400	{object}	i18n.Response
//	@Param		count			query		int		true	"maximum number of results"	minimum(1)	maximum(100000)
//	@Param		offset			query		int		false	"search result offset"		minimum(0)	maximum(100000)
//	@Param		all				query		bool	false	"show all"
//	@Param		q				query		string	false	"search query"
//	@Router		/api/v1/labels [get]
func SearchLabels(router *gin.RouterGroup) {
	router.GET("/labels", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		var f form.SearchLabels

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		result, err := search.Labels(f)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		// TODO c.Header("X-Count", strconv.Itoa(count))
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c, s)

		c.JSON(http.StatusOK, result)
	})
}
