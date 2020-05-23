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
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
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

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		result, err := query.Labels(f)

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

// PUT /api/v1/labels/:uid
func UpdateLabel(router *gin.RouterGroup, conf *config.Config) {
	router.PUT("/labels/:uid", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.Label

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		id := c.Param("uid")
		m, err := query.LabelByUID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrLabelNotFound)
			return
		}

		m.SetName(f.LabelName)
		entity.Db().Save(&m)

		event.Success("label saved")

		PublishLabelEvent(EntityUpdated, id, c)

		c.JSON(http.StatusOK, m)
	})
}

// POST /api/v1/labels/:uid/like
//
// Parameters:
//   uid: string Label UID
func LikeLabel(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/labels/:uid/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		id := c.Param("uid")
		label, err := query.LabelByUID(id)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		label.LabelFavorite = true
		entity.Db().Save(&label)

		if label.LabelPriority < 0 {
			event.Publish("count.labels", event.Data{
				"count": 1,
			})
		}

		PublishLabelEvent(EntityUpdated, id, c)

		c.JSON(http.StatusOK, http.Response{})
	})
}

// DELETE /api/v1/labels/:uid/like
//
// Parameters:
//   uid: string Label UID
func DislikeLabel(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/labels/:uid/like", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		id := c.Param("uid")
		label, err := query.LabelByUID(id)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		label.LabelFavorite = false
		entity.Db().Save(&label)

		if label.LabelPriority < 0 {
			event.Publish("count.labels", event.Data{
				"count": -1,
			})
		}

		PublishLabelEvent(EntityUpdated, id, c)

		c.JSON(http.StatusOK, http.Response{})
	})
}

// GET /api/v1/labels/:uid/thumbnail/:type
//
// Example: /api/v1/labels/cheetah/thumbnail/tile_500
//
// Parameters:
//   uid: string Label UID
//   type: string Thumbnail type, see photoprism.ThumbnailTypes
func LabelThumbnail(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/labels/:uid/thumbnail/:type", func(c *gin.Context) {
		typeName := c.Param("type")
		labelUID := c.Param("uid")
		start := time.Now()

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("label: invalid thumb type %s", txt.Quote(typeName))
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		gc := service.Cache()
		cacheKey := fmt.Sprintf("label-thumbnail:%s:%s", labelUID, typeName)

		if cacheData, ok := gc.Get(cacheKey); ok {
			log.Debugf("cache hit for %s [%s]", cacheKey, time.Since(start))
			c.Data(http.StatusOK, "image/jpeg", cacheData.([]byte))
			return
		}

		f, err := query.LabelThumbByUID(labelUID)

		if err != nil {
			log.Errorf(err.Error())
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		fileName := path.Join(conf.OriginalsPath(), f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("label: could not find original for %s", fileName)
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

		var thumbnail string

		if conf.ThumbUncached() || thumbType.OnDemand() {
			thumbnail, err = thumb.FromFile(fileName, f.FileHash, conf.ThumbPath(), thumbType.Width, thumbType.Height, thumbType.Options...)
		} else {
			thumbnail, err = thumb.FromCache(fileName, f.FileHash, conf.ThumbPath(), thumbType.Width, thumbType.Height, thumbType.Options...)
		}

		if err != nil {
			log.Errorf("label: %s", err)
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		thumbData, err := ioutil.ReadFile(thumbnail)

		if err != nil {
			log.Errorf("label: %s", err)
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		gc.Set(cacheKey, thumbData, time.Hour*4)

		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		c.Data(http.StatusOK, "image/jpeg", thumbData)
	})
}
