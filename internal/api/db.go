package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
)

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

		if f.Table == "photos" {
			result, err := query.TablePhotos(f)
			if err != nil {
				AbortEntityNotFound(c)
				return
			}
			c.JSON(http.StatusOK, result)
		} else if f.Table == "files" {
			result, err := query.TableFiles(f)
			if err != nil {
				AbortEntityNotFound(c)
				return
			}
			c.JSON(http.StatusOK, result)
		} else if f.Table == "albums" {
			result, err := query.TableAlbums(f)
			if err != nil {
				AbortEntityNotFound(c)
				return
			}
			c.JSON(http.StatusOK, result)
		} else if f.Table == "photos_albums" {
			result, err := query.TablePhotosAlbums(f)
			if err != nil {
				AbortEntityNotFound(c)
				return
			}
			c.JSON(http.StatusOK, result)
		} else {

			AbortEntityNotFound(c)
		}
	})
}
