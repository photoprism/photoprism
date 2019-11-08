package api

import (
	"fmt"
	"net/http"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism"
)

// GET /api/v1/thumbnails/:hash/:type
//
// Parameters:
//   hash: string The file hash as returned by the search API
//   type: string Thumbnail type, see photoprism.ThumbnailTypes
func GetThumbnail(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/thumbnails/:hash/:type", func(c *gin.Context) {
		fileHash := c.Param("hash")
		typeName := c.Param("type")

		thumbType, ok := photoprism.ThumbnailTypes[typeName]

		if !ok {
			log.Errorf("invalid type: %s", typeName)
			c.Data(http.StatusBadRequest, "image/svg+xml", photoIconSvg)
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())
		file, err := search.FindFileByHash(fileHash)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		fileName := fmt.Sprintf("%s/%s", conf.OriginalsPath(), file.FileName)

		if !util.Exists(fileName) {
			log.Errorf("could not find original for thumbnail: %s", fileName)
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			file.FileMissing = true
			conf.Db().Save(&file)
			return
		}

		if thumbnail, err := photoprism.ThumbnailFromFile(fileName, file.FileHash, conf.ThumbnailsPath(), thumbType.Width, thumbType.Height, thumbType.Options...); err == nil {
			if c.Query("download") != "" {
				downloadFileName := file.DownloadFileName()

				c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadFileName))
			}

			c.File(thumbnail)
		} else {
			log.Errorf("could not create thumbnail: %s", err)
			c.Data(http.StatusBadRequest, "image/svg+xml", photoIconSvg)
		}
	})
}

// GET /api/v1/labels/:slug/thumbnail/:type
//
// Example: /api/v1/labels/cheetah/thumbnail/tile_500
//
// Parameters:
//   slug: string Label slug name
//   type: string Thumbnail type, see photoprism.ThumbnailTypes
func LabelThumbnail(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/labels/:slug/thumbnail/:type", func(c *gin.Context) {
		typeName := c.Param("type")

		thumbType, ok := photoprism.ThumbnailTypes[typeName]

		if !ok {
			log.Errorf("invalid type: %s", typeName)
			c.Data(http.StatusBadRequest, "image/svg+xml", photoIconSvg)
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())

		// log.Infof("Searching for label slug: %s", c.Param("slug"))

		file, err := search.FindLabelThumbBySlug(c.Param("slug"))

		// log.Infof("Label thumb file: %#v", file)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		fileName := fmt.Sprintf("%s/%s", conf.OriginalsPath(), file.FileName)

		if !util.Exists(fileName) {
			log.Errorf("could not find original for thumbnail: %s", fileName)
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			file.FileMissing = true
			conf.Db().Save(&file)
			return
		}

		if thumbnail, err := photoprism.ThumbnailFromFile(fileName, file.FileHash, conf.ThumbnailsPath(), thumbType.Width, thumbType.Height, thumbType.Options...); err == nil {
			if c.Query("download") != "" {
				downloadFileName := file.DownloadFileName()

				c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadFileName))
			}

			c.File(thumbnail)
		} else {
			log.Errorf("could not create thumbnail: %s", err)
			c.Data(http.StatusBadRequest, "image/svg+xml", photoIconSvg)
		}
	})
}

/* ********** Albums ******** */

func AlbumThumbnail(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/albums/:uuid/thumbnail/:type", func(c *gin.Context) {
		typeName := c.Param("type")
		uuid := c.Param("uuid")

		thumbType, ok := photoprism.ThumbnailTypes[typeName]

		if !ok {
			log.Errorf("invalid type: %s", typeName)
			c.Data(http.StatusBadRequest, "image/svg+xml", photoIconSvg)
			return
		}

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())

		file, err := search.FindAlbumThumbByUUID(uuid)

		if err != nil {
			log.Debugf("album has no photos yet, using generic thumb image: %s", uuid)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		fileName := fmt.Sprintf("%s/%s", conf.OriginalsPath(), file.FileName)

		if !util.Exists(fileName) {
			log.Errorf("could not find original for thumbnail: %s", fileName)
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			file.FileMissing = true
			conf.Db().Save(&file)
			return
		}

		if thumbnail, err := photoprism.ThumbnailFromFile(fileName, file.FileHash, conf.ThumbnailsPath(), thumbType.Width, thumbType.Height, thumbType.Options...); err == nil {
			if c.Query("download") != "" {
				downloadFileName := file.DownloadFileName()

				c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadFileName))
			}

			c.File(thumbnail)
		} else {
			log.Errorf("could not create thumbnail: %s", err)
			c.Data(http.StatusBadRequest, "image/svg+xml", photoIconSvg)
		}
	})
}
