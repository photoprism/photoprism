package api

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
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

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		result, err := query.Albums(f)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		// TODO c.Header("X-Count", strconv.Itoa(count))
		c.Header("X-Limit", strconv.Itoa(f.Count))
		c.Header("X-Offset", strconv.Itoa(f.Offset))

		c.JSON(http.StatusOK, result)
	})
}

// GET /api/v1/albums/:uuid
func GetAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/albums/:uuid", func(c *gin.Context) {
		id := c.Param("uuid")
		m, err := query.AlbumByUUID(id)

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

		m := entity.NewAlbum(f.AlbumName)
		m.AlbumFavorite = f.AlbumFavorite

		log.Debugf("create album: %+v %+v", f, m)

		if res := entity.Db().Create(m); res.Error != nil {
			log.Error(res.Error.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s already exists", txt.Quote(m.AlbumName))})
			return
		}

		event.Success("album created")

		event.Publish("config.updated", event.Data(conf.ClientConfig()))

		PublishAlbumEvent(EntityCreated, m.AlbumUUID, c)

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

		uuid := c.Param("uuid")
		m, err := query.AlbumByUUID(uuid)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		f, err := form.NewAlbum(m)

		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrSaveFailed)
			return
		}

		if err := c.BindJSON(&f); err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, ErrFormInvalid)
			return
		}

		if err := m.Save(f); err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrSaveFailed)
			return
		}

		event.Publish("config.updated", event.Data(conf.ClientConfig()))
		event.Success("album saved")

		PublishAlbumEvent(EntityUpdated, uuid, c)

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

		m, err := query.AlbumByUUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		PublishAlbumEvent(EntityDeleted, id, c)

		conf.Db().Delete(&m)

		event.Publish("config.updated", event.Data(conf.ClientConfig()))
		event.Success(fmt.Sprintf("album %s deleted", txt.Quote(m.AlbumName)))

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
		album, err := query.AlbumByUUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		album.AlbumFavorite = true
		conf.Db().Save(&album)

		event.Publish("config.updated", event.Data(conf.ClientConfig()))
		PublishAlbumEvent(EntityUpdated, id, c)

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
		album, err := query.AlbumByUUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		album.AlbumFavorite = false
		conf.Db().Save(&album)

		event.Publish("config.updated", event.Data(conf.ClientConfig()))
		PublishAlbumEvent(EntityUpdated, id, c)

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

		var f form.Selection

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		uuid := c.Param("uuid")
		a, err := query.AlbumByUUID(uuid)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		photos, err := query.PhotoSelection(f)

		if err != nil {
			log.Errorf("album: %s", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		var added []*entity.PhotoAlbum

		for _, p := range photos {
			added = append(added, entity.NewPhotoAlbum(p.PhotoUUID, a.AlbumUUID).FirstOrCreate())
		}

		if len(added) == 1 {
			event.Success(fmt.Sprintf("one photo added to %s", a.AlbumName))
		} else {
			event.Success(fmt.Sprintf("%d photos added to %s", len(added), a.AlbumName))
		}

		PublishAlbumEvent(EntityUpdated, a.AlbumUUID, c)

		c.JSON(http.StatusOK, gin.H{"message": "photos added to album", "album": a, "added": added})
	})
}

// DELETE /api/v1/albums/:uuid/photos
func RemovePhotosFromAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/albums/:uuid/photos", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

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

		a, err := query.AlbumByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		entity.Db().Where("album_uuid = ? AND photo_uuid IN (?)", a.AlbumUUID, f.Photos).Delete(&entity.PhotoAlbum{})

		event.Success(fmt.Sprintf("photos removed from %s", a.AlbumName))

		PublishAlbumEvent(EntityUpdated, a.AlbumUUID, c)

		c.JSON(http.StatusOK, gin.H{"message": "photos removed from album", "album": a, "photos": f.Photos})
	})
}

// GET /albums/:uuid/download
func DownloadAlbum(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/albums/:uuid/download", func(c *gin.Context) {
		start := time.Now()

		a, err := query.AlbumByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAlbumNotFound)
			return
		}

		p, _, err := query.Photos(form.PhotoSearch{
			Album:  a.AlbumUUID,
			Count:  10000,
			Offset: 0,
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		zipPath := path.Join(conf.TempPath(), "album")
		zipToken := rnd.Token(3)
		zipBaseName := fmt.Sprintf("%s-%s.zip", strings.Title(a.AlbumSlug), zipToken)
		zipFileName := path.Join(zipPath, zipBaseName)

		if err := os.MkdirAll(zipPath, 0700); err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst("failed to create zip folder")})
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
		defer func() { _ = zipWriter.Close() }()

		for _, f := range p {
			fileName := path.Join(conf.OriginalsPath(), f.FileName)
			fileAlias := f.ShareFileName()

			if fs.FileExists(fileName) {
				if err := addFileToZip(zipWriter, fileName, fileAlias); err != nil {
					log.Error(err)
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst("failed to create zip file")})
					return
				}
				log.Infof("album: added %s as %s", txt.Quote(f.FileName), txt.Quote(fileAlias))
			} else {
				log.Warnf("album: %s is missing", txt.Quote(f.FileName))
				f.FileMissing = true
				conf.Db().Save(&f)
			}
		}

		log.Infof("album: archive %s created in %s", txt.Quote(zipBaseName), time.Since(start))
		_ = zipWriter.Close()
		newZipFile.Close()

		if !fs.FileExists(zipFileName) {
			log.Errorf("could not find zip file: %s", zipFileName)
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)
			return
		}

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipBaseName))

		c.File(zipFileName)

		if err := os.Remove(zipFileName); err != nil {
			log.Errorf("album: could not remove %s (%s)", txt.Quote(zipFileName), err.Error())
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
		start := time.Now()

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("album: invalid thumb type %s", typeName)
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
			return
		}

		gc := conf.Cache()
		cacheKey := fmt.Sprintf("album-thumbnail:%s:%s", uuid, typeName)

		if cacheData, ok := gc.Get(cacheKey); ok {
			log.Debugf("album: %s cache hit [%s]", cacheKey, time.Since(start))
			c.Data(http.StatusOK, "image/jpeg", cacheData.([]byte))
			return
		}

		f, err := query.AlbumThumbByUUID(uuid)

		if err != nil {
			log.Debugf("album: no photos yet, using generic image for %s", uuid)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		fileName := path.Join(conf.OriginalsPath(), f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("album: could not find original for %s", fileName)
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			f.FileMissing = true
			entity.Db().Save(&f)
			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsLimit() && c.Query("download") == "" {
			log.Debugf("album: using original, thumbnail size exceeds limit (width %d, height %d)", thumbType.Width, thumbType.Height)
			c.File(fileName)
			return
		}

		var thumbnail string

		if conf.ThumbUncached() || thumbType.OnDemand() {
			thumbnail, err = thumb.FromFile(fileName, f.FileHash, conf.ThumbPath(), thumbType.Width, thumbType.Height, thumbType.Options...)
		} else {
			thumbnail, err = thumb.FromCache(fileName, f.FileHash, conf.ThumbPath(), thumbType.Width, thumbType.Height, thumbType.Options...)
		}

		if err != nil {
			log.Errorf("album: %s", err)
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
			return
		}

		if c.Query("download") != "" {
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.ShareFileName()))
		}

		thumbData, err := ioutil.ReadFile(thumbnail)

		if err != nil {
			log.Errorf("album: %s", err)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		gc.Set(cacheKey, thumbData, time.Hour)

		log.Debugf("album: %s cached [%s]", cacheKey, time.Since(start))

		c.Data(http.StatusOK, "image/jpeg", thumbData)
	})
}
