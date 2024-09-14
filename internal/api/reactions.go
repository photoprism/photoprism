package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/react"
)

// LikePhoto flags a photo as favorite.
//
//	@Summary	flags a photo as favorite
//	@Id			LikePhoto
//	@Tags		Photos
//	@Produce	json
//	@Success	200				{object}	gin.H
//	@Failure	401,403,404,500	{object}	i18n.Response
//	@Param		uid				path		string	true	"photo uid"
//	@Router		/api/v1/photos/{uid}/like [post]
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

			SaveSidecarYaml(&m)
			PublishPhotoEvent(StatusUpdated, id, c)
		}

		c.JSON(http.StatusOK, gin.H{"photo": m})
	})
}

// DislikePhoto removes the favorite flags from a photo.
//
//	@Summary	removes the favorite flags from a photo
//	@Id			DislikePhoto
//	@Tags		Photos
//	@Produce	json
//	@Success	200				{object}	gin.H
//	@Failure	401,403,404,500	{object}	i18n.Response
//	@Param		uid				path		string	true	"photo uid"
//	@Router		/api/v1/photos/{uid}/like [delete]
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

			SaveSidecarYaml(&m)
			PublishPhotoEvent(StatusUpdated, id, c)
		}

		c.JSON(http.StatusOK, gin.H{"photo": m})
	})
}
