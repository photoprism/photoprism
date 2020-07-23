package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
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
//   favorite:  bool   Find favorites only
func GetPhotos(router *gin.RouterGroup) {
	router.GET("/photos", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.PhotoSearch

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		// Guests may only see public content in shared albums.
		if s.Guest() {
			if f.Album == "" || !s.HasShare(f.Album) {
				AbortUnauthorized(c)
				return
			}

			f.Public = true
			f.Private = false
			f.Hidden = false
			f.Archived = false
			f.Review = false
		}

		result, count, err := query.PhotoSearch(f)

		if err != nil {
			log.Error(err)
			AbortBadRequest(c)
			return
		}

		c.Header("X-Count", strconv.Itoa(count))
		c.Header("X-Limit", strconv.Itoa(f.Count))
		c.Header("X-Offset", strconv.Itoa(f.Offset))

		c.JSON(http.StatusOK, result)
	})
}
