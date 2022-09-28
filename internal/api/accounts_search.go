package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/internal/service"
)

// SearchAccounts finds accounts and returns them as JSON.
//
// GET /api/v1/accounts
func SearchAccounts(router *gin.RouterGroup) {
	router.GET("/accounts", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAccounts, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		conf := service.Config()

		if conf.Demo() || conf.DisableSettings() {
			c.JSON(http.StatusOK, entity.Accounts{})
			return
		}

		var f form.SearchAccounts

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
