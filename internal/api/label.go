package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
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

// GET /api/v1/labels/:uid/t/:token/:type
//
// Parameters:
//   uid: string Label UID
//   type: string Thumbnail type, see photoprism.ThumbnailTypes
func LabelThumbnail(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/labels/:uid/t/:token/:type", func(c *gin.Context) {
		if InvalidToken(c, conf) {
			c.Data(http.StatusForbidden, "image/svg+xml", labelIconSvg)
			return
		}

		start := time.Now()
		typeName := c.Param("type")
		uid := c.Param("uid")

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("label-thumbnail: invalid type %s", txt.Quote(typeName))
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		cache := service.Cache()
		cacheKey := fmt.Sprintf("label-thumbnail:%s:%s", uid, typeName)

		if cacheData, err := cache.Get(cacheKey); err == nil {
			log.Debugf("cache hit for %s [%s]", cacheKey, time.Since(start))

			var cached ThumbCache

			if err := json.Unmarshal(cacheData, &cached); err != nil {
				log.Errorf("label-thumbnail: %s not found", uid)
				c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
				return
			}

			if !fs.FileExists(cached.FileName) {
				log.Errorf("label-thumbnail: %s not found", uid)
				c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
				return
			}

			if c.Query("download") != "" {
				c.FileAttachment(cached.FileName, cached.ShareName)
			} else {
				c.File(cached.FileName)
			}

			return
		}

		f, err := query.LabelThumbByUID(uid)

		if err != nil {
			log.Errorf(err.Error())
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		fileName := path.Join(conf.OriginalsPath(), f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("label-thumbnail: file %s is missing", txt.Quote(f.FileName))
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logError("label-thumbnail", f.Update("FileMissing", true))

			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsLimit() {
			log.Debugf("label-thumbnail: using original, size exceeds limit (width %d, height %d)", thumbType.Width, thumbType.Height)

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
			log.Errorf("label-thumbnail: %s", err)
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("label-thumbnail: %s has empty thumb name - bug?", filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		if cached, err := json.Marshal(ThumbCache{thumbnail, f.ShareFileName()}); err == nil {
			logError("label-thumbnail", cache.Set(cacheKey, cached))
			log.Debugf("cached %s [%s]", cacheKey, time.Since(start))
		}

		if c.Query("download") != "" {
			c.FileAttachment(thumbnail, f.ShareFileName())
		} else {
			c.File(thumbnail)
		}
	})
}
