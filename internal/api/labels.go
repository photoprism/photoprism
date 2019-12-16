package api

import (
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/repo"
	"github.com/photoprism/photoprism/internal/util"
)

// GET /api/v1/labels
func GetLabels(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/labels", func(c *gin.Context) {
		var f form.LabelSearch

		r := repo.New(conf.OriginalsPath(), conf.Db())
		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		result, err := r.Labels(f)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		c.Header("X-Result-Count", strconv.Itoa(f.Count))
		c.Header("X-Result-Offset", strconv.Itoa(f.Offset))

		c.JSON(http.StatusOK, result)
	})
}

// POST /api/v1/labels/:slug/like
//
// Parameters:
//   slug: string Label slug name
func LikeLabel(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/labels/:slug/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		r := repo.New(conf.OriginalsPath(), conf.Db())

		label, err := r.FindLabelBySlug(c.Param("slug"))

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		label.LabelFavorite = true
		conf.Db().Save(&label)

		c.JSON(http.StatusOK, http.Response{})
	})
}

// DELETE /api/v1/labels/:slug/like
//
// Parameters:
//   slug: string Label slug name
func DislikeLabel(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/labels/:slug/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		r := repo.New(conf.OriginalsPath(), conf.Db())

		label, err := r.FindLabelBySlug(c.Param("slug"))

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		label.LabelFavorite = false
		conf.Db().Save(&label)

		c.JSON(http.StatusOK, http.Response{})
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

		r := repo.New(conf.OriginalsPath(), conf.Db())

		// log.Infof("Searching for label slug: %s", c.Param("slug"))

		file, err := r.FindLabelThumbBySlug(c.Param("slug"))

		// log.Infof("Label thumb file: %#v", file)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		fileName := path.Join(conf.OriginalsPath(), file.FileName)

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
