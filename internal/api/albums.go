package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/forms"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/util"
)

// GET /api/v1/albums
func GetAlbums(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/albums", func(c *gin.Context) {
		var form forms.AlbumSearchForm

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())
		err := c.MustBindWith(&form, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		result, err := search.Albums(form)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		c.Header("X-Result-Count", strconv.Itoa(form.Count))
		c.Header("X-Result-Offset", strconv.Itoa(form.Offset))

		c.JSON(http.StatusOK, result)
	})
}

type AlbumParams struct {
	AlbumName string `json:"AlbumName"`
}

// POST /api/v1/albums
func CreateAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/albums", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var params AlbumParams

		if err := c.BindJSON(&params); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		m := models.NewAlbum(params.AlbumName)

		if res := conf.Db().Create(m); res.Error != nil {
			log.Error(res.Error.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("\"%s\" already exists", m.AlbumName)})
			return
		}

		event.Success(fmt.Sprintf("Album %s created", m.AlbumName))

		c.JSON(http.StatusOK, m)
	})
}

// PUT /api/v1/albums/:uuid
func UpdateAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.PUT("/albums/:uuid", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var params AlbumParams

		if err := c.BindJSON(&params); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		id := c.Param("uuid")
		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())

		m, err := search.FindAlbumByUUID(id)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		m.Rename(params.AlbumName)
		conf.Db().Save(&m)

		event.Publish("config.updated", event.Data(conf.ClientConfig()))
		event.Success(fmt.Sprintf("Album %s updated", m.AlbumName))

		c.JSON(http.StatusOK, m)
	})
}

// POST /api/v1/albums/:uuid/like
//
// Parameters:
//   uuid: string Album UUID
func LikeAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/albums/:uuid/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())

		album, err := search.FindAlbumByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		album.AlbumFavorite = true
		conf.Db().Save(&album)

		event.Publish("config.updated", event.Data(conf.ClientConfig()))

		c.JSON(http.StatusOK, http.Response{})
	})
}

// DELETE /api/v1/albums/:uuid/like
//
// Parameters:
//   uuid: string Album UUID
func DislikeAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/albums/:uuid/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())

		album, err := search.FindAlbumByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		album.AlbumFavorite = false
		conf.Db().Save(&album)

		event.Publish("config.updated", event.Data(conf.ClientConfig()))

		c.JSON(http.StatusOK, http.Response{})
	})
}

// POST /api/v1/albums/:uuid/photos
func AddPhotosToAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/albums/:uuid/photos", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var params PhotoUUIDs

		if err := c.BindJSON(&params); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		if len(params.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no photos selected")})
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())
		a, err := search.FindAlbumByUUID(c.Param("uuid"))

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

// DELETE /api/v1/albums/:uuid/photos
func RemovePhotosFromAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/albums/:uuid/photos", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var params PhotoUUIDs

		if err := c.BindJSON(&params); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		if len(params.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst("no photos selected")})
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())
		a, err := search.FindAlbumByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		log.Infof("adding %d photos to album %s", len(params.Photos), a.AlbumName)

		db := conf.Db()

		db.Where("album_uuid = ? AND photo_uuid IN (?)", a.AlbumUUID, params.Photos).Delete(&models.PhotoAlbum{})

		event.Success(fmt.Sprintf("photos removed from %s", a.AlbumName))

		c.JSON(http.StatusOK, gin.H{"message": "photos removed from album", "album": a, "photos": params.Photos})
	})
}
