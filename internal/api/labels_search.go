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
// GET /api/v1/labels
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
