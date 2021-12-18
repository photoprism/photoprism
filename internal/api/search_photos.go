package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/search"
)

// SearchPhotos searches the pictures index and returns the result as JSON.
//
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
func SearchPhotos(router *gin.RouterGroup) {
	router.GET("/photos", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.SearchPhotos

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

			f.UID = ""
			f.Public = true
			f.Private = false
			f.Hidden = false
			f.Archived = false
			f.Review = false
		}

		result, count, err := search.Photos(f)

		if err != nil {
			log.Warnf("search: %s", err)
			AbortBadRequest(c)
			return
		}

		AddCountHeader(c, count)
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c)

		c.JSON(http.StatusOK, result)
	})
}
