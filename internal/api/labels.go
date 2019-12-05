package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/util"
)

// GET /api/v1/labels
func GetLabels(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/labels", func(c *gin.Context) {
		var f form.LabelSearch

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())
		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		result, err := search.Labels(f)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		c.Header("X-Result-Count", strconv.Itoa(f.Count))
		c.Header("X-Result-Offset", strconv.Itoa(f.Offset))

		c.JSON(http.StatusOK, result)
	})
}

// POST /api/v1/labels/:slug/like
//
// Parameters:
//   slug: string Label slug name
func LikeLabel(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/labels/:slug/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())

		label, err := search.FindLabelBySlug(c.Param("slug"))

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		label.LabelFavorite = true
		conf.Db().Save(&label)

		c.JSON(http.StatusOK, http.Response{})
	})
}

// DELETE /api/v1/labels/:slug/like
//
// Parameters:
//   slug: string Label slug name
func DislikeLabel(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/labels/:slug/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())

		label, err := search.FindLabelBySlug(c.Param("slug"))

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		label.LabelFavorite = false
		conf.Db().Save(&label)

		c.JSON(http.StatusOK, http.Response{})
	})
}
