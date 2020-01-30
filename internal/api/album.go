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
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GET /api/v1/albums
func GetAlbums(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/albums", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.AlbumSearch

		q := query.New(conf.OriginalsPath(), conf.Db())
		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		result, err := q.Albums(f)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
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
		q := query.New(conf.OriginalsPath(), conf.Db())
		m, err := q.FindAlbumByUUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		q := query.New(conf.OriginalsPath(), conf.Db())
		m := entity.NewAlbum(f.AlbumName)
		m.AlbumFavorite = f.AlbumFavorite

		log.Debugf("create album: %+v %+v", f, m)

		if res := conf.Db().Create(m); res.Error != nil {
			log.Error(res.Error.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("\"%s\" already exists", m.AlbumName)})
			return
		}

		/* TODO: Not needed if we send config.updated
		event.Publish("count.albums", event.Data{
			"count": 1,
		})
		*/

		event.Success(fmt.Sprintf("album \"%s\" created", m.AlbumName))

		event.Publish("config.updated", event.Data(conf.ClientConfig()))

		PublishAlbumEvent(EntityCreated, m.AlbumUUID, c, q)

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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		id := c.Param("uuid")
		q := query.New(conf.OriginalsPath(), conf.Db())

		m, err := q.FindAlbumByUUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		m.Rename(f.AlbumName)
		conf.Db().Save(&m)

		event.Publish("config.updated", event.Data(conf.ClientConfig()))
		event.Success(fmt.Sprintf("album \"%s\" saved", m.AlbumName))

		PublishAlbumEvent(EntityUpdated, id, c, q)

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
		q := query.New(conf.OriginalsPath(), conf.Db())

		m, err := q.FindAlbumByUUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		PublishAlbumEvent(EntityDeleted, id, c, q)

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

		id := c.Param("uuid")
		q := query.New(conf.OriginalsPath(), conf.Db())

		album, err := q.FindAlbumByUUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		album.AlbumFavorite = true
		conf.Db().Save(&album)

		event.Publish("config.updated", event.Data(conf.ClientConfig()))
		PublishAlbumEvent(EntityUpdated, id, c, q)

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

		id := c.Param("uuid")
		q := query.New(conf.OriginalsPath(), conf.Db())
		album, err := q.FindAlbumByUUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		album.AlbumFavorite = false
		conf.Db().Save(&album)

		event.Publish("config.updated", event.Data(conf.ClientConfig()))
		PublishAlbumEvent(EntityUpdated, id, c, q)

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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if len(f.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst("no photos selected")})
			return
		}

		q := query.New(conf.OriginalsPath(), conf.Db())
		a, err := q.FindAlbumByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		db := conf.Db()
		var added []*entity.PhotoAlbum
		var failed []string

		for _, photoUUID := range f.Photos {
			if p, err := q.FindPhotoByUUID(photoUUID); err != nil {
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
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if len(f.Photos) == 0 {
			log.Error("no photos selected")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst("no photos selected")})
			return
		}

		q := query.New(conf.OriginalsPath(), conf.Db())
		a, err := q.FindAlbumByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
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

		q := query.New(conf.OriginalsPath(), conf.Db())
		a, err := q.FindAlbumByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		p, err := q.Photos(form.PhotoSearch{
			Album:  a.AlbumUUID,
			Count:  10000,
			Offset: 0,
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		zipPath := path.Join(conf.ExportPath(), "album")
		zipToken := rnd.Token(3)
		zipBaseName := fmt.Sprintf("%s-%s.zip", strings.Title(a.AlbumSlug), zipToken)
		zipFileName := path.Join(zipPath, zipBaseName)

		if err := os.MkdirAll(zipPath, 0700); err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst("failed to create zip directory")})
			return
		}

		newZipFile, err := os.Create(zipFileName)

		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		defer newZipFile.Close()

		zipWriter := zip.NewWriter(newZipFile)
		defer zipWriter.Close()

		for _, f := range p {
			fileName := path.Join(conf.OriginalsPath(), f.FileName)
			fileAlias := f.DownloadFileName()

			if fs.FileExists(fileName) {
				if err := addFileToZip(zipWriter, fileName, fileAlias); err != nil {
					log.Error(err)
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst("failed to create zip file")})
					return
				}
				log.Infof("album: added \"%s\" as \"%s\"", f.FileName, fileAlias)
			} else {
				log.Warnf("album: \"%s\" is missing", f.FileName)
				f.FileMissing = true
				conf.Db().Save(&f)
			}
		}

		log.Infof("album: archive \"%s\" created in %s", zipBaseName, time.Since(start))

		zipWriter.Close()
		newZipFile.Close()

		if !fs.FileExists(zipFileName) {
			log.Errorf("could not find zip file: %s", zipFileName)
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)
			return
		}

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipBaseName))

		c.File(zipFileName)

		if err := os.Remove(zipFileName); err != nil {
			log.Errorf("album: could not remove \"%s\" %s", zipFileName, err.Error())
		}
	})
}

// GET /api/v1/albums/:uuid/thumbnail/:type
//
// Parameters:
//   uuid: string Album UUID
//   type: string Thumbnail type, see photoprism.ThumbnailTypes
func AlbumThumbnail(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/albums/:uuid/thumbnail/:type", func(c *gin.Context) {
		typeName := c.Param("type")
		uuid := c.Param("uuid")

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("thumbs: invalid type \"%s\"", typeName)
			c.Data(http.StatusBadRequest, "image/svg+xml", photoIconSvg)
			return
		}

		q := query.New(conf.OriginalsPath(), conf.Db())

		f, err := q.FindAlbumThumbByUUID(uuid)

		if err != nil {
			log.Debugf("album has no photos yet, using generic thumb image: %s", uuid)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		fileName := path.Join(conf.OriginalsPath(), f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("could not find original for thumbnail: %s", fileName)
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			f.FileMissing = true
			conf.Db().Save(&f)
			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsLimit() && c.Query("download") == "" {
			log.Debugf("album: using original, thumbnail size exceeds limit (width %d, height %d)", thumbType.Width, thumbType.Height)

			c.File(fileName)

			return
		}

		if thumbnail, err := thumb.FromFile(fileName, f.FileHash, conf.ThumbnailsPath(), thumbType.Width, thumbType.Height, thumbType.Options...); err == nil {
			if c.Query("download") != "" {
				c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.DownloadFileName()))
			}

			c.File(thumbnail)
		} else {
			log.Errorf("could not create thumbnail: %s", err)
			c.Data(http.StatusBadRequest, "image/svg+xml", photoIconSvg)
		}
	})
}
