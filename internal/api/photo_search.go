package api

import (
	"net/http"
	"strconv"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/form"
)

// GET /api/v1/photos
//
// Query:
//   q:         string Query string
//   label:     string Label
//   cat:       string Category
//   country:   string Country code
//   camera:    int    UpdateCamera ID
//   order:     string Sort order
//   count:     int    Max result count (required)
//   offset:    int    Result offset
//   before:    date   Find photos taken before (format: "2006-01-02")
//   after:     date   Find photos taken after (format: "2006-01-02")
//   favorites: bool   Find favorites only
func GetPhotos(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/photos", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.PhotoSearch

		q := query.New(conf.OriginalsPath(), conf.Db())
		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		result, err := q.Photos(f)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		c.Header("X-Result-Count", strconv.Itoa(f.Count))
		c.Header("X-Result-Offset", strconv.Itoa(f.Offset))

		c.JSON(http.StatusOK, result)
	})
}
