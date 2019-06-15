package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/util"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

type BatchParams struct {
	Ids []int `json:"ids"`
}

// POST /api/v1/batch/photos/delete
func BatchPhotosDelete(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/batch/photos/delete", func(c *gin.Context) {
		start := time.Now()

		var params BatchParams

		if err := c.BindJSON(&params); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
		}

		if len(params.Ids) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no photos selected")})
		}

		log.Infof("deleting photos: %#v", params.Ids)

		db := conf.Db()

		db.Where("id IN (?)", params.Ids).Delete(&models.Photo{})
		db.Where("photo_id IN (?)", params.Ids).Delete(&models.File{})

		elapsed := time.Since(start)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("photos deleted in %s", elapsed)})
	})
}

// POST /api/v1/batch/photos/private
func BatchPhotosPrivate(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/batch/photos/private", func(c *gin.Context) {
		start := time.Now()

		var params BatchParams

		if err := c.BindJSON(&params); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
		}

		if len(params.Ids) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no photos selected")})
		}

		log.Infof("marking photos as private: %#v", params.Ids)

		db := conf.Db()

		db.Model(models.Photo{}).Where("id IN (?)", params.Ids).Updates(models.Photo{PhotoPrivate: true})

		elapsed := time.Since(start)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("photos marked as private in %s", elapsed)})
	})
}
