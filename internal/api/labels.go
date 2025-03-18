package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/txt"
)

// UpdateLabel updates label properties.
//
// PUT /api/v1/labels/:uid
//
//	@Summary	updates label name
//	@Id			UpdateLabel
//	@Tags		Labels
//	@Accept		json
//	@Produce	json
//	@Success	200				{object}	entity.Label
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		uid				path		string		true	"Label UID"
//	@Param		label			body		form.Label	true	"Label Name"
//	@Router		/api/v1/labels/{uid} [put]
func UpdateLabel(router *gin.RouterGroup) {
	router.PUT("/labels/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		// Find label by UID.
		id := clean.UID(c.Param("uid"))
		m, err := query.LabelByUID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrLabelNotFound)
			return
		}

		// Create new label form.
		frm, frmErr := form.NewLabel(m)

		if frmErr != nil {
			Abort(c, http.StatusBadRequest, i18n.ErrBadRequest)
			return
		}

		// Set form values from request.
		if frmErr = c.BindJSON(frm); frmErr != nil {
			AbortBadRequest(c)
			return
		} else if frmErr = frm.Validate(); frmErr != nil {
			AbortInvalidName(c)
			return
		}

		// Save label and return new model values if successful.
		if err = m.SaveForm(frm); err != nil {
			log.Errorf("label: %s", clean.Error(err))
			AbortSaveFailed(c)
			return
		}

		event.SuccessMsg(i18n.MsgLabelSaved)

		PublishLabelEvent(StatusUpdated, id, c)

		c.JSON(http.StatusOK, m)
	})
}

// LikeLabel flags a label as favorite.
//
//	@Summary	sets favorite flag for a label
//	@Id			LikeLabel
//	@Tags		Labels
//	@Accept		json
//	@Produce	json
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		uid				path		string	true	"Label UID"
//	@Router		/api/v1/labels/{uid}/like [post]
func LikeLabel(router *gin.RouterGroup) {
	router.POST("/labels/:uid/like", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		id := clean.UID(c.Param("uid"))
		label, err := query.LabelByUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		if err := label.Update("LabelFavorite", true); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		if label.LabelPriority < 0 {
			event.Publish("count.labels", event.Data{
				"count": 1,
			})
		}

		PublishLabelEvent(StatusUpdated, id, c)

		c.JSON(http.StatusOK, http.Response{})
	})
}

// DislikeLabel removes the favorite flag from a label.
//
//	@Summary	removes favorite flag from a label
//	@Id			DislikeLabel
//	@Tags		Labels
//	@Accept		json
//	@Produce	json
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		uid				path		string	true	"Label UID"
//	@Router		/api/v1/labels/{uid}/like [delete]
func DislikeLabel(router *gin.RouterGroup) {
	router.DELETE("/labels/:uid/like", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		id := clean.UID(c.Param("uid"))
		label, err := query.LabelByUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		if err = label.Update("LabelFavorite", false); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		if label.LabelPriority < 0 {
			event.Publish("count.labels", event.Data{
				"count": -1,
			})
		}

		PublishLabelEvent(StatusUpdated, id, c)

		c.JSON(http.StatusOK, http.Response{})
	})
}
