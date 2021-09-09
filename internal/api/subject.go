package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GetSubjects finds and returns subjects as JSON.
//
// GET /api/v1/subjects
func GetSubjects(router *gin.RouterGroup) {
	router.GET("/subjects", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceSubjects, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.SubjectSearch

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		result, err := query.SubjectSearch(f)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		AddCountHeader(c, len(result))
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c)

		c.JSON(http.StatusOK, result)
	})
}

// GetSubject returns a subject as JSON.
//
// GET /api/v1/subjects/:uid
func GetSubject(router *gin.RouterGroup) {
	router.GET("/subjects/:uid", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceSubjects, acl.ActionRead)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		if subj := entity.FindSubject(c.Param("uid")); subj == nil {
			Abort(c, http.StatusNotFound, i18n.ErrSubjectNotFound)
			return
		} else {
			c.JSON(http.StatusOK, subj)
		}

	})
}
