package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GetSubject returns a subject as JSON.
//
//	@Summary	returns a subject as JSON
//	@Id			GetSubject
//	@Tags		Subjects
//	@Produce	json
//	@Success	200				{object}	entity.Subject
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		uid				path		string	true	"subject uid"
//	@Router		/api/v1/subjects/{uid} [get]
func GetSubject(router *gin.RouterGroup) {
	router.GET("/subjects/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePeople, acl.ActionView)

		if s.Abort(c) {
			return
		}

		if subj := entity.FindSubject(clean.UID(c.Param("uid"))); subj == nil {
			Abort(c, http.StatusNotFound, i18n.ErrSubjectNotFound)
			return
		} else {
			c.JSON(http.StatusOK, subj)
		}
	})
}

// UpdateSubject updates subject properties.
//
//	@Summary	updates subject properties
//	@Id			UpdateSubject
//	@Tags		Subjects
//	@Produce	json
//	@Success	200						{object}	entity.Subject
//	@Failure	400,401,403,404,429,500	{object}	i18n.Response
//	@Param		uid						path		string			true	"subject uid"
//	@Param		subject					body		form.Subject	true	"properties to be updated (only submit values that should be changed)"
//	@Router		/api/v1/subjects/{uid} [put]
func UpdateSubject(router *gin.RouterGroup) {
	router.PUT("/subjects/:uid", func(c *gin.Context) {
		if err := mutex.UpdatePeople.Start(); err != nil {
			AbortBusy(c)
			return
		}

		defer mutex.UpdatePeople.Stop()

		s := Auth(c, acl.ResourcePeople, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))
		m := entity.FindSubject(uid)

		if m == nil {
			Abort(c, http.StatusNotFound, i18n.ErrSubjectNotFound)
			return
		}

		// Create request value form.
		f, err := form.NewSubject(*m)

		// Assign and validate request form values.
		if err != nil {
			log.Errorf("subject: %s (new form)", err)
			AbortSaveFailed(c)
			return
		} else if err = c.BindJSON(&f); err != nil {
			log.Errorf("subject: %s (update form)", err)
			AbortBadRequest(c)
			return
		}

		// Update subject from form values.
		if changed, err := m.SaveForm(f); err != nil {
			log.Errorf("subject: %s", err)
			AbortSaveFailed(c)
			return
		} else if changed {
			// Show success message.
			if m.IsPerson() {
				event.SuccessMsg(i18n.MsgPersonSaved)
			} else {
				event.SuccessMsg(i18n.MsgSubjectSaved)
			}
		}

		c.JSON(http.StatusOK, m)
	})
}

// LikeSubject flags a subject as favorite.
//
//	@Summary	flags a subject as favorite
//	@Id			LikeSubject
//	@Tags		Subjects
//	@Produce	json
//	@Failure	401,403,404,429,500	{object}	i18n.Response
//	@Param		uid					path		string	true	"subject uid"
//	@Router		/api/v1/subjects/{uid}/like [post]
func LikeSubject(router *gin.RouterGroup) {
	router.POST("/subjects/:uid/like", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePeople, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))
		subj := entity.FindSubject(uid)

		if subj == nil {
			Abort(c, http.StatusNotFound, i18n.ErrSubjectNotFound)
			return
		}

		if err := subj.Update("SubjFavorite", true); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		PublishSubjectEvent(StatusUpdated, uid, c)

		c.JSON(http.StatusOK, http.Response{})
	})
}

// DislikeSubject removes the favorite flag from a subject.
//
//	@Summary	removes the favorite flag from a subject
//	@Id			DislikeSubject
//	@Tags		Subjects
//	@Produce	json
//	@Failure	401,403,404,429,500	{object}	i18n.Response
//	@Param		uid					path		string	true	"subject uid"
//	@Router		/api/v1/subjects/{uid}/like [delete]
func DislikeSubject(router *gin.RouterGroup) {
	router.DELETE("/subjects/:uid/like", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePeople, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))
		subj := entity.FindSubject(uid)

		if subj == nil {
			Abort(c, http.StatusNotFound, i18n.ErrSubjectNotFound)
			return
		}

		if err := subj.Update("SubjFavorite", false); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		PublishSubjectEvent(StatusUpdated, uid, c)

		c.JSON(http.StatusOK, http.Response{})
	})
}
