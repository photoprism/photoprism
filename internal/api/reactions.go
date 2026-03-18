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

// ratingRequest is used to parse the rating value from a JSON request body.
type ratingRequest struct {
	Rating int `json:"Rating"`
}

// LikePhoto flags a photo as favorite.
//
//	@Summary	flags a photo as favorite
//	@Id			LikePhoto
//	@Tags		Photos
//	@Accept		json
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

		if get.Config().Develop() && acl.Rules.Allow(acl.ResourcePhotos, s.GetUserRole(), acl.ActionReact) {
			logWarn("react", m.React(s.GetUser(), react.Find("love")))
		}

		if acl.Rules.Allow(acl.ResourcePhotos, s.GetUserRole(), acl.ActionUpdate) {
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
//	@Accept		json
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

		if get.Config().Develop() && acl.Rules.Allow(acl.ResourcePhotos, s.GetUserRole(), acl.ActionReact) {
			logWarn("react", m.UnReact(s.GetUser()))
		}

		if acl.Rules.Allow(acl.ResourcePhotos, s.GetUserRole(), acl.ActionUpdate) {
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

// RatePhoto sets the star rating (0-5) for a photo.
//
//	@Summary	sets the star rating (0-5) for a photo
//	@Id			RatePhoto
//	@Tags		Photos
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	gin.H
//	@Failure	400,401,403,404,500	{object}	i18n.Response
//	@Param		uid					path		string			true	"photo uid"
//	@Param		rating				body		ratingRequest	true	"star rating 0-5"
//	@Router		/api/v1/photos/{uid}/rating [post]
func RatePhoto(router *gin.RouterGroup) {
	router.POST("/photos/:uid/rating", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		id := clean.UID(c.Param("uid"))
		m, err := query.PhotoByUID(id)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		var req ratingRequest

		if err = c.BindJSON(&req); err != nil {
			AbortBadRequest(c)
			return
		}

		if err = m.SetRating(req.Rating); err != nil {
			log.Errorf("photo: %s", err.Error())
			AbortSaveFailed(c)
			return
		}

		SaveSidecarYaml(&m)
		PublishPhotoEvent(StatusUpdated, id, c)

		c.JSON(http.StatusOK, gin.H{"photo": m})
	})
}
