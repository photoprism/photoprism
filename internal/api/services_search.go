package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// SearchServices finds account settings and returns them as JSON.
//
//	@Summary	finds services and returns them as JSON
//	@Id			SearchServices
//	@Tags		Services
//	@Produce	json
//	@Success	200				{object}	entity.Services
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		count			query		int	true	"maximum number of results"	minimum(1)	maximum(100000)
//	@Router		/api/v1/services [get]
func SearchServices(router *gin.RouterGroup) {
	router.GET("/services", func(c *gin.Context) {
		s := Auth(c, acl.ResourceServices, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if conf.Demo() || conf.DisableSettings() {
			c.JSON(http.StatusOK, entity.Services{})
			return
		}

		var frm form.SearchServices

		err := c.MustBindWith(&frm, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		result, err := search.Accounts(frm)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		// TODO c.Header("X-Count", strconv.Itoa(count))
		AddLimitHeader(c, frm.Count)
		AddOffsetHeader(c, frm.Offset)

		c.JSON(http.StatusOK, result)
	})
}
