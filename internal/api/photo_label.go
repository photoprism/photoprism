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

		q := query.New(conf.OriginalsPath(), conf.Db())
		m, err := q.FindPhotoByUUID(c.Param("uuid"))
		db := conf.Db()

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		var f form.Label

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		lm := entity.NewLabel(f.LabelName, f.LabelPriority).FirstOrCreate(db)

		if lm.New && f.LabelPriority >= 0 {
			event.Publish("count.labels", event.Data{
				"count": 1,
			})
		}

		plm := entity.NewPhotoLabel(m.ID, lm.ID, f.LabelUncertainty, "manual").FirstOrCreate(db)

		if plm.LabelUncertainty > f.LabelUncertainty {
			plm.LabelUncertainty = f.LabelUncertainty
			plm.LabelSource = "manual"

			if err := db.Save(&plm).Error; err != nil {
				log.Errorf("label: %s", err)
			}
		}

		db.Save(&lm)

		p, err := q.PreloadPhotoByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

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

		q := query.New(conf.OriginalsPath(), conf.Db())
		m, err := q.FindPhotoByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		labelId, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		db := conf.Db()
		db.Where("photo_id = ? AND label_id = ?", m.ID, labelId).Delete(&entity.PhotoLabel{})

		p, err := q.PreloadPhotoByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		c.JSON(http.StatusOK, p)
	})
}
