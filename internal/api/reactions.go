package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/react"
)

// LikePhoto flags a photo as favorite.
//
// POST /api/v1/photos/:uid/like
func LikePhoto(router *gin.RouterGroup) {
	router.POST("/photos/:uid/like", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourcePhotos, acl.Permissions{acl.ActionUpdate, acl.ActionReact})

		if s.Abort(c) {
			return
		}

		id := clean.UID(c.Param("uid"))
		m, err := query.PhotoByUID(id)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if get.Config().Experimental() && acl.Rules.Allow(acl.ResourcePhotos, s.UserRole(), acl.ActionReact) {
			logWarn("react", m.React(s.User(), react.Find("love")))
		}

		if acl.Rules.Allow(acl.ResourcePhotos, s.UserRole(), acl.ActionUpdate) {
			err = m.SetFavorite(true)

			if err != nil {
				log.Errorf("photo: %s", err.Error())
				AbortSaveFailed(c)
				return
			}

			SavePhotoAsYaml(&m)
			PublishPhotoEvent(StatusUpdated, id, c)
		}

		c.JSON(http.StatusOK, gin.H{"photo": m})
	})
}

// DislikePhoto removes the favorite flags from a photo.
//
// DELETE /api/v1/photos/:uid/like
func DislikePhoto(router *gin.RouterGroup) {
	router.DELETE("/photos/:uid/like", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourcePhotos, acl.Permissions{acl.ActionUpdate, acl.ActionReact})

		if s.Abort(c) {
			return
		}

		id := clean.UID(c.Param("uid"))
		m, err := query.PhotoByUID(id)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if get.Config().Experimental() && acl.Rules.Allow(acl.ResourcePhotos, s.UserRole(), acl.ActionReact) {
			logWarn("react", m.UnReact(s.User()))
		}

		if acl.Rules.Allow(acl.ResourcePhotos, s.UserRole(), acl.ActionUpdate) {
			err = m.SetFavorite(false)

			if err != nil {
				log.Errorf("photo: %s", err.Error())
				AbortSaveFailed(c)
				return
			}

			SavePhotoAsYaml(&m)
			PublishPhotoEvent(StatusUpdated, id, c)
		}

		c.JSON(http.StatusOK, gin.H{"photo": m})
	})
}
