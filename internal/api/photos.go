package api

import (
	"net/http"
	"strconv"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/forms"
	"github.com/photoprism/photoprism/internal/photoprism"
)

// GET /api/v1/photos
//
// Query:
//   q:         string Query string
//   label:     string Label
//   cat:       string Category
//   country:   string Country code
//   camera:    int    Camera ID
//   order:     string Sort order
//   count:     int    Max result count (required)
//   offset:    int    Result offset
//   before:    date   Find photos taken before (format: "2006-01-02")
//   after:     date   Find photos taken after (format: "2006-01-02")
//   favorites: bool   Find favorites only
func GetPhotos(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/photos", func(c *gin.Context) {
		var form forms.PhotoSearchForm

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())
		err := c.MustBindWith(&form, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		result, err := search.Photos(form)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		c.Header("X-Result-Count", strconv.Itoa(form.Count))
		c.Header("X-Result-Offset", strconv.Itoa(form.Offset))

		c.JSON(http.StatusOK, result)
	})
}

// POST /api/v1/photos/:id/like
//
// Parameters:
//   id: int Photo ID as returned by the API
func LikePhoto(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/photos/:id/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())
		photoID, err := strconv.ParseUint(c.Param("id"), 10, 64)

		if err != nil {
			log.Errorf("could not find image for id: %s", err.Error())
			c.Data(http.StatusNotFound, "image", []byte(""))
			return
		}

		photo, err := search.FindPhotoByID(photoID)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		photo.PhotoFavorite = true
		conf.Db().Save(&photo)

		c.JSON(http.StatusOK, http.Response{})
	})
}

// DELETE /api/v1/photos/:photoId/like
//
// Parameters:
//   id: int Photo ID as returned by the API
func DislikePhoto(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/photos/:id/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)

		if err != nil {
			log.Errorf("could not find image for id: %s", err.Error())
			c.Data(http.StatusNotFound, "image", []byte(""))
			return
		}

		photo, err := search.FindPhotoByID(id)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		photo.PhotoFavorite = false
		conf.Db().Save(&photo)

		c.JSON(http.StatusOK, http.Response{})
	})
}
