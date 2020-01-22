package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GET /api/v1/labels
func GetLabels(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/labels", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.LabelSearch

		q := query.New(conf.OriginalsPath(), conf.Db())
		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		result, err := q.Labels(f)
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		c.Header("X-Result-Count", strconv.Itoa(f.Count))
		c.Header("X-Result-Offset", strconv.Itoa(f.Offset))

		c.JSON(http.StatusOK, result)
	})
}

// POST /api/v1/labels/:uuid/like
//
// Parameters:
//   uuid: string Label UUID
func LikeLabel(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/labels/:uuid/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		q := query.New(conf.OriginalsPath(), conf.Db())

		label, err := q.FindLabelByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		label.LabelFavorite = true
		conf.Db().Save(&label)

		if label.LabelPriority < 0 {
			event.Publish("count.labels", event.Data{
				"count": 1,
			})
		}

		c.JSON(http.StatusOK, http.Response{})
	})
}

// DELETE /api/v1/labels/:uuid/like
//
// Parameters:
//   uuid: string Label UUID
func DislikeLabel(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/labels/:uuid/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		q := query.New(conf.OriginalsPath(), conf.Db())

		label, err := q.FindLabelByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		label.LabelFavorite = false
		conf.Db().Save(&label)

		if label.LabelPriority < 0 {
			event.Publish("count.labels", event.Data{
				"count": -1,
			})
		}

		c.JSON(http.StatusOK, http.Response{})
	})
}

// GET /api/v1/labels/:uuid/thumbnail/:type
//
// Example: /api/v1/labels/cheetah/thumbnail/tile_500
//
// Parameters:
//   uuid: string Label UUID
//   type: string Thumbnail type, see photoprism.ThumbnailTypes
func LabelThumbnail(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/labels/:uuid/thumbnail/:type", func(c *gin.Context) {
		typeName := c.Param("type")
		labelUUID := c.Param("uuid")
		start := time.Now()

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("thumbs: invalid type \"%s\"", typeName)
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		q := query.New(conf.OriginalsPath(), conf.Db())

		gc := conf.Cache()
		cacheKey := fmt.Sprintf("label-thumbnail:%s:%s", labelUUID, typeName)

		if cacheData, ok := gc.Get(cacheKey); ok {
			log.Debugf("%s cache hit [%s]", cacheKey, time.Since(start))
			c.Data(http.StatusOK, "image/jpeg", cacheData.([]byte))
			return
		}

		f, err := q.FindLabelThumbByUUID(labelUUID)

		if err != nil {
			log.Errorf(err.Error())
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		fileName := path.Join(conf.OriginalsPath(), f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("could not find original for thumbnail: %s", fileName)
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			f.FileMissing = true
			conf.Db().Save(&f)
			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsLimit() {
			log.Debugf("label: using original, thumbnail size exceeds limit (width %d, height %d)", thumbType.Width, thumbType.Height)

			c.File(fileName)

			return
		}

		if thumbnail, err := thumb.FromFile(fileName, f.FileHash, conf.ThumbnailsPath(), thumbType.Width, thumbType.Height, thumbType.Options...); err == nil {
			thumbData, err := ioutil.ReadFile(thumbnail)

			if err != nil {
				log.Errorf("could not read thumbnail: %s", err)
				c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
				return
			}

			gc.Set(cacheKey, thumbData, time.Hour*4)

			log.Debugf("%s cached [%s]", cacheKey, time.Since(start))

			c.Data(http.StatusOK, "image/jpeg", thumbData)
		} else {
			log.Errorf("could not create thumbnail: %s", err)

			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}
	})
}
