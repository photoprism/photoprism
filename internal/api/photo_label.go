package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
)

// POST /api/v1/photos/:uuid/label
//
// Parameters:
//   uuid: string PhotoUUID as returned by the API
func AddPhotoLabel(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/photos/:uuid/label", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		m, err := query.PhotoByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		var f form.Label

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		lm := entity.NewLabel(f.LabelName, f.LabelPriority).FirstOrCreate()

		if lm.New && f.LabelPriority >= 0 {
			event.Publish("count.labels", event.Data{
				"count": 1,
			})
		}

		plm := entity.NewPhotoLabel(m.ID, lm.ID, f.Uncertainty, "manual").FirstOrCreate()

		if plm.Uncertainty > f.Uncertainty {
			plm.Uncertainty = f.Uncertainty
			plm.LabelSrc = entity.SrcManual

			if err := entity.Db().Save(&plm).Error; err != nil {
				log.Errorf("label: %s", err)
			}
		}

		entity.Db().Save(&lm)

		p, err := query.PreloadPhotoByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		if err := p.Save(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		PublishPhotoEvent(EntityUpdated, c.Param("uuid"), c)

		event.Success("label updated")

		c.JSON(http.StatusOK, p)
	})
}

// DELETE /api/v1/photos/:uuid/label/:id
//
// Parameters:
//   uuid: string PhotoUUID as returned by the API
//   id: int LabelId as returned by the API
func RemovePhotoLabel(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/photos/:uuid/label/:id", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		m, err := query.PhotoByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		labelId, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		label, err := query.PhotoLabel(m.ID, uint(labelId))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if label.LabelSrc == entity.SrcManual {
			entity.Db().Delete(&label)
		} else {
			label.Uncertainty = 100
			entity.Db().Save(&label)
		}

		p, err := query.PreloadPhotoByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		if err := p.Save(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		PublishPhotoEvent(EntityUpdated, c.Param("uuid"), c)

		event.Success("label removed")

		c.JSON(http.StatusOK, p)
	})
}

// PUT /api/v1/photos/:uuid/label/:id
//
// Parameters:
//   uuid: string PhotoUUID as returned by the API
//   id: int LabelId as returned by the API
func UpdatePhotoLabel(router *gin.RouterGroup, conf *config.Config) {
	router.PUT("/photos/:uuid/label/:id", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		// TODO: Code clean-up, simplify

		m, err := query.PhotoByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		labelId, err := strconv.Atoi(c.Param("id"))

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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if err := label.Save(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		p, err := query.PreloadPhotoByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		if err := p.Save(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		PublishPhotoEvent(EntityUpdated, c.Param("uuid"), c)

		event.Success("label saved")

		c.JSON(http.StatusOK, p)
	})
}
