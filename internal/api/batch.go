package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/util"

	"github.com/gin-gonic/gin"
)

type BatchParams struct {
	Photos []int `json:"photos"`
}

// POST /api/v1/batch/photos/delete
func BatchPhotosDelete(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/batch/photos/delete", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		start := time.Now()

		var params BatchParams

		if err := c.BindJSON(&params); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		if len(params.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no photos selected")})
			return
		}

		log.Infof("deleting photos: %#v", params.Photos)

		db := conf.Db()

		db.Where("photo_uuid IN (?)", params.Photos).Delete(&models.Photo{})

		elapsed := time.Since(start)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("photos deleted in %s", elapsed)})
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

		var params BatchParams

		if err := c.BindJSON(&params); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		if len(params.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no photos selected")})
			return
		}

		log.Infof("marking photos as private: %#v", params.Photos)

		db := conf.Db()

		db.Model(models.Photo{}).Where("photo_uuid IN (?)", params.Photos).UpdateColumn("photo_private", gorm.Expr("IF (`photo_private`, 0, 1)"))

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

		var params BatchParams

		if err := c.BindJSON(&params); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		if len(params.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no photos selected")})
			return
		}

		log.Infof("marking photos as story: %#v", params.Photos)

		db := conf.Db()

		db.Model(models.Photo{}).Where("photo_uuid IN (?)", params.Photos).Updates(map[string]interface{}{
			"photo_story":   gorm.Expr("IF (`photo_story`, 0, 1)"),
		})

		elapsed := time.Since(start)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("photos marked as story in %s", elapsed)})
	})
}

type BatchPhotosAlbumParams struct {
	Photos    []string `json:"photos"`
	AlbumUUID string   `json:"album"`
}

// POST /api/v1/batch/photos/album
func BatchPhotosAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/batch/photos/album", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var params BatchPhotosAlbumParams

		if err := c.BindJSON(&params); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		if params.AlbumUUID == "" {
			log.Error("no album selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no album selected")})
			return
		}

		if len(params.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no photos selected")})
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())
		a, err := search.FindAlbumByUUID(params.AlbumUUID)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		log.Infof("adding %d photos to album %s", len(params.Photos), a.AlbumName)

		db := conf.Db()
		var added []*models.PhotoAlbum
		var failed []string

		for _, photoUUID := range params.Photos {
			if p, err := search.FindPhotoByUUID(photoUUID); err != nil {
				failed = append(failed, photoUUID)
			} else {
				added = append(added, models.NewPhotoAlbum(p.PhotoUUID, a.AlbumUUID).FirstOrCreate(db))
			}
		}

		if len(added) == 1 {
			event.Success(fmt.Sprintf("one photo added to %s", a.AlbumName))
		} else {
			event.Success(fmt.Sprintf("%d photos added to %s", len(added), a.AlbumName))
		}

		c.JSON(http.StatusOK, gin.H{"message": "photos added to album", "album": a, "added": added, "failed": failed})
	})
}
