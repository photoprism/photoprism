package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/forms"
	"github.com/photoprism/photoprism/internal/photoprism"
	"log"
	"net/http"
	"strconv"
)

// GetPhotos searches the databse for photos based on a form.
func GetPhotos(router *gin.RouterGroup, conf *photoprism.Config) {
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

func LikePhoto(router *gin.RouterGroup, conf *photoprism.Config) {
	router.POST("/photos/:photoId/like", func(c *gin.Context) {
		search := photoprism.NewSearch(conf.OriginalsPath, conf.GetDb())

		photoId, err := strconv.ParseUint(c.Param("photoId"), 10, 64)

		if err == nil {
			photo := search.FindPhotoById(photoId)
			photo.PhotoFavorite = true
			conf.GetDb().Save(&photo)
			c.JSON(http.StatusAccepted, http.Response{})
		} else {
			log.Printf("could not find image for id: %s", err.Error())
			c.Data(http.StatusNotFound, "image", []byte(""))
		}
	})
}

func DislikePhoto(router *gin.RouterGroup, conf *photoprism.Config) {
	router.DELETE("/photos/:photoId/like", func(c *gin.Context) {
		search := photoprism.NewSearch(conf.OriginalsPath, conf.GetDb())

		photoId, err := strconv.ParseUint(c.Param("photoId"), 10, 64)

		if err == nil {
			photo := search.FindPhotoById(photoId)
			photo.PhotoFavorite = false
			conf.GetDb().Save(&photo)
			c.JSON(http.StatusAccepted, http.Response{})
		} else {
			log.Printf("could not find image for id: %s", err.Error())
			c.Data(http.StatusNotFound, "image", []byte(""))
		}
	})
}
