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
	"github.com/photoprism/photoprism/internal/util"

	"github.com/gin-gonic/gin"
)

// POST /api/v1/batch/photos/delete
func BatchPhotosDelete(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/batch/photos/delete", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		start := time.Now()

		var f form.PhotoUUIDs

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		if len(f.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no photos selected")})
			return
		}

		log.Infof("deleting photos: %#v", f.Photos)

		db := conf.Db()

		db.Where("photo_uuid IN (?)", f.Photos).Delete(&entity.Photo{})

		elapsed := int(time.Since(start).Seconds())

		event.Publish("count.photos", event.Data{
			"count": len(f.Photos) * -1,
		})

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("photos deleted in %d s", elapsed)})
	})
}

// POST /api/v1/batch/albums/delete
func BatchAlbumsDelete(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/batch/albums/delete", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.AlbumUUIDs

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		if len(f.Albums) == 0 {
			log.Error("no albums selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no albums selected")})
			return
		}

		log.Infof("deleting albums: %#v", f.Albums)

		db := conf.Db()

		db.Where("album_uuid IN (?)", f.Albums).Delete(&entity.Album{})
		db.Where("album_uuid IN (?)", f.Albums).Delete(&entity.PhotoAlbum{})

		event.Publish("config.updated", event.Data(conf.ClientConfig()))

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

		var f form.PhotoUUIDs

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		if len(f.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no photos selected")})
			return
		}

		log.Infof("marking photos as private: %#v", f.Photos)

		db := conf.Db()

		db.Model(entity.Photo{}).Where("photo_uuid IN (?)", f.Photos).UpdateColumn("photo_private", gorm.Expr("IF (`photo_private`, 0, 1)"))

		elapsed := time.Since(start)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("photos marked as private in %s", elapsed)})
	})
}

// POST /api/v1/batch/photos/story
func BatchPhotosStory(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/batch/photos/story", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		start := time.Now()

		var f form.PhotoUUIDs

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		if len(f.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no photos selected")})
			return
		}

		log.Infof("marking photos as story: %#v", f.Photos)

		db := conf.Db()

		db.Model(entity.Photo{}).Where("photo_uuid IN (?)", f.Photos).Updates(map[string]interface{}{
			"photo_story": gorm.Expr("IF (`photo_story`, 0, 1)"),
		})

		elapsed := time.Since(start)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("photos marked as story in %s", elapsed)})
	})
}
