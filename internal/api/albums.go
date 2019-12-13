package api

import (
	"archive/zip"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/repo"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"
)

// GET /api/v1/albums
func GetAlbums(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/albums", func(c *gin.Context) {
		var f form.AlbumSearch

		r := repo.New(conf.OriginalsPath(), conf.Db())
		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		result, err := r.Albums(f)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		c.Header("X-Result-Count", strconv.Itoa(f.Count))
		c.Header("X-Result-Offset", strconv.Itoa(f.Offset))

		c.JSON(http.StatusOK, result)
	})
}

// GET /api/v1/albums/:uuid
func GetAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/albums/:uuid", func(c *gin.Context) {
		id := c.Param("uuid")
		r := repo.New(conf.OriginalsPath(), conf.Db())
		m, err := r.FindAlbumByUUID(id)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		c.JSON(http.StatusOK, m)
	})
}

// POST /api/v1/albums
func CreateAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/albums", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.Album

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		m := entity.NewAlbum(f.AlbumName)

		if res := conf.Db().Create(m); res.Error != nil {
			log.Error(res.Error.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("\"%s\" already exists", m.AlbumName)})
			return
		}

		event.Publish("count.albums", event.Data{
			"count": 1,
		})

		event.Success(fmt.Sprintf("album \"%s\" created", m.AlbumName))

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

		var f form.Album

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		id := c.Param("uuid")
		r := repo.New(conf.OriginalsPath(), conf.Db())

		m, err := r.FindAlbumByUUID(id)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		m.Rename(f.AlbumName)
		conf.Db().Save(&m)

		event.Publish("config.updated", event.Data(conf.ClientConfig()))
		event.Success(fmt.Sprintf("album \"%s\" updated", m.AlbumName))

		c.JSON(http.StatusOK, m)
	})
}

// DELETE /api/v1/albums/:uuid
func DeleteAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/albums/:uuid", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		id := c.Param("uuid")
		r := repo.New(conf.OriginalsPath(), conf.Db())

		m, err := r.FindAlbumByUUID(id)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		conf.Db().Delete(&m)

		event.Publish("config.updated", event.Data(conf.ClientConfig()))
		event.Success(fmt.Sprintf("album \"%s\" deleted", m.AlbumName))

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

		r := repo.New(conf.OriginalsPath(), conf.Db())

		album, err := r.FindAlbumByUUID(c.Param("uuid"))

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

		r := repo.New(conf.OriginalsPath(), conf.Db())
		album, err := r.FindAlbumByUUID(c.Param("uuid"))

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

		r := repo.New(conf.OriginalsPath(), conf.Db())
		a, err := r.FindAlbumByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		db := conf.Db()
		var added []*entity.PhotoAlbum
		var failed []string

		for _, photoUUID := range f.Photos {
			if p, err := r.FindPhotoByUUID(photoUUID); err != nil {
				failed = append(failed, photoUUID)
			} else {
				added = append(added, entity.NewPhotoAlbum(p.PhotoUUID, a.AlbumUUID).FirstOrCreate(db))
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

		r := repo.New(conf.OriginalsPath(), conf.Db())
		a, err := r.FindAlbumByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		db := conf.Db()

		db.Where("album_uuid = ? AND photo_uuid IN (?)", a.AlbumUUID, f.Photos).Delete(&entity.PhotoAlbum{})

		event.Success(fmt.Sprintf("photos removed from %s", a.AlbumName))

		c.JSON(http.StatusOK, gin.H{"message": "photos removed from album", "album": a, "photos": f.Photos})
	})
}

// GET /albums/:uuid/download
func DownloadAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/albums/:uuid/download", func(c *gin.Context) {

		start := time.Now()

		r := repo.New(conf.OriginalsPath(), conf.Db())
		a, err := r.FindAlbumByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		p, err := r.Photos(form.PhotoSearch{
			Album:  a.AlbumUUID,
			Count:  10000,
			Offset: 0,
		})

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		zipPath := path.Join(conf.ExportPath(), "album")
		zipToken := util.RandomToken(3)
		zipBaseName := fmt.Sprintf("%s-%s.zip", strings.Title(a.AlbumSlug), zipToken)
		zipFileName := path.Join(zipPath, zipBaseName)

		if err := os.MkdirAll(zipPath, 0700); err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": util.UcFirst("failed to create zip directory")})
			return
		}

		newZipFile, err := os.Create(zipFileName)

		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		defer newZipFile.Close()

		zipWriter := zip.NewWriter(newZipFile)
		defer zipWriter.Close()

		for _, file := range p {
			fileName := path.Join(conf.OriginalsPath(), file.FileName)
			fileAlias := file.DownloadFileName()

			if util.Exists(fileName) {
				if err := addFileToZip(zipWriter, fileName, fileAlias); err != nil {
					log.Error(err)
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": util.UcFirst("failed to create zip file")})
					return
				}
				log.Infof("album: added \"%s\" as \"%s\"", file.FileName, fileAlias)
			} else {
				log.Warnf("album: \"%s\" is missing", file.FileName)
				file.FileMissing = true
				conf.Db().Save(&file)
			}
		}

		log.Infof("album: archive \"%s\" created in %s", zipBaseName, time.Since(start))

		zipWriter.Close()
		newZipFile.Close()

		if !util.Exists(zipFileName) {
			log.Errorf("could not find zip file: %s", zipFileName)
			c.Data(404, "image/svg+xml", photoIconSvg)
			return
		}

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipBaseName))

		c.File(zipFileName)

		if err := os.Remove(zipFileName); err != nil {
			log.Errorf("album: could not remove \"%s\" %s", zipFileName, err.Error())
		}
	})
}
