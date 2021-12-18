package api

import (
	"net/http"
	"strconv"

	"github.com/photoprism/photoprism/pkg/sanitize"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
)

// POST /api/v1/photos/:uid/label
//
// Parameters:
//   uid: string PhotoUID as returned by the API
func AddPhotoLabel(router *gin.RouterGroup) {
	router.POST("/photos/:uid/label", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		m, err := query.PhotoByUID(sanitize.IdString(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		var f form.Label

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		labelEntity := entity.FirstOrCreateLabel(entity.NewLabel(f.LabelName, f.LabelPriority))

		if labelEntity == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed creating label"})
			return
		}

		if err := labelEntity.Restore(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not restore label"})
		}

		photoLabel := entity.FirstOrCreatePhotoLabel(entity.NewPhotoLabel(m.ID, labelEntity.ID, f.Uncertainty, "manual"))

		if photoLabel == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed updating photo label"})
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

		p, err := query.PhotoPreloadByUID(sanitize.IdString(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if err := p.SaveLabels(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		PublishPhotoEvent(EntityUpdated, c.Param("uid"), c)

		event.Success("label updated")

		c.JSON(http.StatusOK, p)
	})
}

// DELETE /api/v1/photos/:uid/label/:id
//
// Parameters:
//   uid: string PhotoUID as returned by the API
//   id: int LabelId as returned by the API
func RemovePhotoLabel(router *gin.RouterGroup) {
	router.DELETE("/photos/:uid/label/:id", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		m, err := query.PhotoByUID(sanitize.IdString(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		labelId, err := strconv.Atoi(sanitize.Token(c.Param("id")))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		label, err := query.PhotoLabel(m.ID, uint(labelId))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if label.LabelSrc == classify.SrcManual || label.LabelSrc == classify.SrcKeyword {
			logError("label", entity.Db().Delete(&label).Error)
		} else {
			label.Uncertainty = 100
			logError("label", entity.Db().Save(&label).Error)
		}

		p, err := query.PhotoPreloadByUID(sanitize.IdString(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		logError("label", p.RemoveKeyword(label.Label.LabelName))

		if err := p.SaveLabels(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		PublishPhotoEvent(EntityUpdated, sanitize.IdString(c.Param("uid")), c)

		event.Success("label removed")

		c.JSON(http.StatusOK, p)
	})
}

// PUT /api/v1/photos/:uid/label/:id
//
// Parameters:
//   uid: string PhotoUID as returned by the API
//   id: int LabelId as returned by the API
func UpdatePhotoLabel(router *gin.RouterGroup) {
	router.PUT("/photos/:uid/label/:id", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		// TODO: Code clean-up, simplify

		m, err := query.PhotoByUID(sanitize.IdString(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		labelId, err := strconv.Atoi(sanitize.Token(c.Param("id")))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		label, err := query.PhotoLabel(m.ID, uint(labelId))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if err := c.BindJSON(&label); err != nil {
			AbortBadRequest(c)
			return
		}

		if err := label.Save(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		p, err := query.PhotoPreloadByUID(sanitize.IdString(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if err := p.SaveLabels(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		PublishPhotoEvent(EntityUpdated, sanitize.IdString(c.Param("uid")), c)

		event.Success("label saved")

		c.JSON(http.StatusOK, p)
	})
}
