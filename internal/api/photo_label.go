package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// AddPhotoLabel adds a label to a photo.
//
// The request parameters are:
//
//   - uid: string PhotoUID as returned by the API
//
// POST /api/v1/photos/:uid/label
func AddPhotoLabel(router *gin.RouterGroup) {
	router.POST("/photos/:uid/label", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		m, err := query.PhotoByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		var f form.Label

		// Assign and validate request form values.
		if err = c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		labelEntity := entity.FirstOrCreateLabel(entity.NewLabel(f.LabelName, f.LabelPriority))

		if labelEntity == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create label"})
			return
		}

		if err = labelEntity.Restore(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not restore label"})
			return
		}

		photoLabel := entity.FirstOrCreatePhotoLabel(entity.NewPhotoLabel(m.ID, labelEntity.ID, f.Uncertainty, "manual"))

		if photoLabel == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to update photo label"})
			return
		}

		if photoLabel.Uncertainty > f.Uncertainty {
			if err := photoLabel.Updates(map[string]interface{}{
				"Uncertainty": f.Uncertainty,
				"LabelSrc":    entity.SrcManual,
			}); err != nil {
				log.Errorf("label: %s", err)
			}
		}

		p, err := query.PhotoPreloadByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if err = p.SaveLabels(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		PublishPhotoEvent(StatusUpdated, c.Param("uid"), c)

		event.Success("label updated")

		c.JSON(http.StatusOK, p)
	})
}

// RemovePhotoLabel removes a label from a photo.
//
// The request parameters are:
//
//   - uid: string PhotoUID as returned by the API
//   - id: int LabelId as returned by the API
//
// DELETE /api/v1/photos/:uid/label/:id
func RemovePhotoLabel(router *gin.RouterGroup) {
	router.DELETE("/photos/:uid/label/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		m, err := query.PhotoByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		labelId, err := strconv.Atoi(clean.Token(c.Param("id")))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		label, err := query.PhotoLabel(m.ID, uint(labelId))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		if label.LabelSrc == classify.SrcManual || label.LabelSrc == classify.SrcKeyword {
			logErr("label", entity.Db().Delete(&label).Error)
		} else {
			label.Uncertainty = 100
			logErr("label", entity.Db().Save(&label).Error)
		}

		p, err := query.PhotoPreloadByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		logErr("label", p.RemoveKeyword(label.Label.LabelName))

		if err := p.SaveLabels(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		PublishPhotoEvent(StatusUpdated, clean.UID(c.Param("uid")), c)

		event.Success("label removed")

		c.JSON(http.StatusOK, p)
	})
}

// UpdatePhotoLabel changes a photo labels.
//
// The request parameters are:
//
//   - uid: string PhotoUID as returned by the API
//   - id: int LabelId as returned by the API
//
// PUT /api/v1/photos/:uid/label/:id
func UpdatePhotoLabel(router *gin.RouterGroup) {
	router.PUT("/photos/:uid/label/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		// TODO: Code clean-up, simplify

		m, err := query.PhotoByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		labelId, err := strconv.Atoi(clean.Token(c.Param("id")))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		label, err := query.PhotoLabel(m.ID, uint(labelId))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		if err := c.BindJSON(&label); err != nil {
			AbortBadRequest(c)
			return
		}

		if err := label.Save(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		p, err := query.PhotoPreloadByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if err := p.SaveLabels(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		PublishPhotoEvent(StatusUpdated, clean.UID(c.Param("uid")), c)

		event.Success("label saved")

		c.JSON(http.StatusOK, p)
	})
}
