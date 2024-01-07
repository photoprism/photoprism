package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// UpdateLabel updates label properties.
//
// PUT /api/v1/labels/:uid
func UpdateLabel(router *gin.RouterGroup) {
	router.PUT("/labels/:uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLabels, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		var f form.Label

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		id := clean.UID(c.Param("uid"))
		m, err := query.LabelByUID(id)

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrLabelNotFound)
			return
		}

		m.SetName(f.LabelName)
		entity.Db().Save(&m)

		event.SuccessMsg(i18n.MsgLabelSaved)

		PublishLabelEvent(StatusUpdated, id, c)

		c.JSON(http.StatusOK, m)
	})
}

// LikeLabel flags a label as favorite.
//
// Request Parameters:
// - uid: string Label UID
//
// POST /api/v1/labels/:uid/like
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
// Request Parameters:
// - uid: string Label UID
//
// DELETE /api/v1/labels/:uid/like
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

		if err := label.Update("LabelFavorite", false); err != nil {
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
