package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
)

// SearchServices finds account settings and returns them as JSON.
//
// GET /api/v1/services
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

		var f form.SearchServices

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		result, err := search.Accounts(f)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		// TODO c.Header("X-Count", strconv.Itoa(count))
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)

		c.JSON(http.StatusOK, result)
	})
}
