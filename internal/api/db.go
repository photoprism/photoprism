package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
)

type dbResult struct {
	QueryTimestamp time.Time
	Results        interface{}
}

// GET /api/v1/db
func GetTable(router *gin.RouterGroup) {
	router.GET("/db", func(c *gin.Context) {
		s := Auth(c, acl.ResourceDefault, acl.ActionView)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.DbSearch

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		var result dbResult
		result.QueryTimestamp = time.Now()

		if f.Table == "photos" {
			result.Results, err = query.TablePhotos(f)
			if err != nil {
				AbortEntityNotFound(c)
				return
			}
		} else if f.Table == "files" {
			result.Results, err = query.TableFiles(f)
			if err != nil {
				AbortEntityNotFound(c)
				return
			}
		} else if f.Table == "albums" {
			result.Results, err = query.TableAlbums(f)
			if err != nil {
				AbortEntityNotFound(c)
				return
			}
		} else if f.Table == "photos_albums" {
			result.Results, err = query.TablePhotosAlbums(f)
			if err != nil {
				AbortEntityNotFound(c)
				return
			}
		} else {
			AbortEntityNotFound(c)
			return
		}
		c.JSON(http.StatusOK, result)
	})
}
