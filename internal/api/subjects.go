package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
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
// GET /api/v1/subjects/:uid
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
// PUT /api/v1/subjects/:uid
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

		// Initialize form.
		f, err := form.NewSubject(*m)

		if err != nil {
			log.Errorf("subject: %s (new form)", err)
			AbortSaveFailed(c)
			return
		} else if err := c.BindJSON(&f); err != nil {
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
// The request parameters are:
//
//   - uid: string Subject UID
//
// POST /api/v1/subjects/:uid/like
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
// The request parameters are:
//
//   - uid: string Subject UID
//
// DELETE /api/v1/subjects/:uid/like
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
