package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GetFace returns a face as JSON.
//
//	@Summary	returns a face as JSON
//	@Id			GetFace
//	@Tags		Faces
//	@Produce	json
//	@Success	200				{object}	entity.Face
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		id				path		string	true	"face id"
//	@Router		/api/v1/faces/{id} [get]
func GetFace(router *gin.RouterGroup) {
	router.GET("/faces/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePeople, acl.ActionView)

		// Abort if permission was not granted.
		if s.Abort(c) {
			return
		}

		f := form.SearchFaces{UID: c.Param("id"), Markers: true}

		if results, err := search.Faces(f); err != nil || len(results) < 1 {
			Abort(c, http.StatusNotFound, i18n.ErrFaceNotFound)
			return
		} else {
			c.JSON(http.StatusOK, results[0])
		}
	})
}

// UpdateFace updates face properties.
//
//	@Summary	updates face properties
//	@Id			UpdateFace
//	@Tags		Faces
//	@Produce	json
//	@Success	200						{object}	entity.Face
//	@Failure	400,401,403,404,429,500	{object}	i18n.Response
//	@Param		id						path		string		true	"face id"
//	@Param		face					body		form.Face	true	"properties to be updated"
//	@Router		/api/v1/faces/{id} [put]
func UpdateFace(router *gin.RouterGroup) {
	router.PUT("/faces/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePeople, acl.ActionUpdate)

		// Abort if permission was not granted.
		if s.Abort(c) {
			return
		}

		var f form.Face

		// Assign and validate request form values.
		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		faceId := clean.Token(c.Param("id"))
		m := entity.FindFace(faceId)

		if m == nil {
			Abort(c, http.StatusNotFound, i18n.ErrFaceNotFound)
			return
		}

		// Change visibility?
		if !f.FaceHidden && f.FaceHidden == m.FaceHidden {
			// Do nothing.
		} else if err := m.Update("FaceHidden", f.FaceHidden); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		// Change subject?
		if f.SubjUID == "" {
			// Do nothing.
		} else if err := m.SetSubjectUID(f.SubjUID); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		event.SuccessMsg(i18n.MsgChangesSaved)

		c.JSON(http.StatusOK, m)
	})
}
