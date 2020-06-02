package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/gin-gonic/gin"
)

// POST /api/v1/batch/photos/archive
func BatchPhotosArchive(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/batch/photos/archive", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		start := time.Now()

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if len(f.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst("no photos selected")})
			return
		}

		log.Infof("photos: archiving %#v", f.Photos)

		err := entity.Db().Where("photo_uid IN (?)", f.Photos).Delete(&entity.Photo{}).Error

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrSaveFailed)
			return
		}

		if err := entity.UpdatePhotoCounts(); err != nil {
			log.Errorf("photos: %s", err)
		}

		elapsed := int(time.Since(start).Seconds())

		UpdateClientConfig(conf)

		event.EntitiesArchived("photos", f.Photos)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("photos archived in %d s", elapsed)})
	})
}

// POST /api/v1/batch/photos/restore
func BatchPhotosRestore(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/batch/photos/restore", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		start := time.Now()

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if len(f.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst("no photos selected")})
			return
		}

		log.Infof("restoring photos: %#v", f.Photos)

		err := entity.Db().Unscoped().Model(&entity.Photo{}).Where("photo_uid IN (?)", f.Photos).
			UpdateColumn("deleted_at", gorm.Expr("NULL")).Error

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrSaveFailed)
			return
		}

		if err := entity.UpdatePhotoCounts(); err != nil {
			log.Errorf("photos: %s", err)
		}

		elapsed := int(time.Since(start).Seconds())

		UpdateClientConfig(conf)

		event.EntitiesRestored("photos", f.Photos)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("photos restored in %d s", elapsed)})
	})
}

// POST /api/v1/batch/albums/delete
func BatchAlbumsDelete(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/batch/albums/delete", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if len(f.Albums) == 0 {
			log.Error("no albums selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst("no albums selected")})
			return
		}

		log.Infof("albums: deleting %#v", f.Albums)

		entity.Db().Where("album_uid IN (?)", f.Albums).Delete(&entity.Album{})
		entity.Db().Where("album_uid IN (?)", f.Albums).Delete(&entity.PhotoAlbum{})

		UpdateClientConfig(conf)

		event.EntitiesDeleted("albums", f.Albums)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("albums deleted")})
	})
}

// POST /api/v1/batch/photos/private
func BatchPhotosPrivate(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/batch/photos/private", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		start := time.Now()

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if len(f.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst("no photos selected")})
			return
		}

		log.Infof("marking photos as private: %#v", f.Photos)

		err := entity.Db().Model(entity.Photo{}).Where("photo_uid IN (?)", f.Photos).UpdateColumn("photo_private",
			gorm.Expr("CASE WHEN photo_private > 0 THEN 0 ELSE 1 END")).Error

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrSaveFailed)
			return
		}

		if err := entity.UpdatePhotoCounts(); err != nil {
			log.Errorf("photos: %s", err)
		}

		if entities, err := query.PhotoSelection(f); err == nil {
			event.EntitiesUpdated("photos", entities)
		}

		UpdateClientConfig(conf)

		elapsed := time.Since(start)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("photos marked as private in %s", elapsed)})
	})
}

// POST /api/v1/batch/labels/delete
func BatchLabelsDelete(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/batch/labels/delete", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if len(f.Labels) == 0 {
			log.Error("no labels selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst("no labels selected")})
			return
		}

		log.Infof("labels: deleting %#v", f.Labels)

		var labels entity.Labels

		if err := entity.Db().Where("label_uid IN (?)", f.Labels).Find(&labels).Error; err != nil {
			logError("labels", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrDeleteFailed)
			return
		}

		for _, label := range labels {
			logError("labels", label.Delete())
		}

		UpdateClientConfig(conf)

		event.EntitiesDeleted("labels", f.Labels)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("labels deleted")})
	})
}
