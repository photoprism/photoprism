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

// SearchLenses finds and returns lenses as JSON.
//
//	@Summary	finds and returns lenses as JSON
//	@Id			SearchLenses
//	@Tags		Lenses
//	@Produce	json
//	@Success	200				{array}		search.Lens
//	@Failure	401,429,403,400	{object}	i18n.Response
//	@Param		count			query		int		true	"maximum number of results"	minimum(1)	maximum(100000)
//	@Param		offset			query		int		false	"search result offset"		minimum(0)	maximum(100000)
//	@Param		nomake			query		bool	false	"show where make is blank"
//	@Param		q				query		string	false	"search query"
//	@Router		/api/v1/lenses [get]
func SearchLenses(router *gin.RouterGroup) {
	router.GET("/lenses", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLenses, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		var frm form.SearchLenses

		err := c.MustBindWith(&frm, binding.Form)

		if err != nil {
			AbortBadRequest(c, err)
			return
		}

		// Search matching lenses.
		result, err := search.Lenses(frm)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		// TODO c.Header("X-Count", strconv.Itoa(count))
		AddLimitHeader(c, frm.Count)
		AddOffsetHeader(c, frm.Offset)
		AddTokenHeaders(c, s)

		c.JSON(http.StatusOK, result)
	})
}
