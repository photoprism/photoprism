package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/forms"
	"github.com/photoprism/photoprism/internal/photoprism"
	"net/http"
	"strconv"
)

func GetPhotos (router *gin.RouterGroup, conf *photoprism.Config) {
	router.GET("/photos", func(c *gin.Context) {
		var form forms.PhotoSearchForm

		search := photoprism.NewSearch(conf.OriginalsPath, conf.GetDb())

		c.MustBindWith(&form, binding.Form)

		result, err := search.Photos(form)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		}

		c.Header("x-result-count", strconv.Itoa(form.Count))
		c.Header("x-result-offset", strconv.Itoa(form.Offset))

		c.JSON(http.StatusOK, result)
	})
}