package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/txt"
)

// POST /api/v1/index
func StartIndexing(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/index", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		start := time.Now()

		var f form.IndexOptions

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		path := conf.OriginalsPath()

		event.Info(fmt.Sprintf("indexing photos in \"%s\"", filepath.Base(path)))

		cancel := func(err error) {
			log.Error(err.Error())
			event.Publish("index.completed", event.Data{"path": path, "seconds": int(time.Since(start).Seconds())})
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		}

		if f.ConvertRaw && !conf.ReadOnly() {
			convert := service.Convert()

			if err := convert.Start(conf.OriginalsPath()); err != nil {
				cancel(err)
				return
			}
		}

		if f.CreateThumbs {
			rs := service.Resample()

			if err := rs.Start(false); err != nil {
				cancel(err)
				return
			}
		}

		ind := service.Index()

		if f.CompleteRescan {
			ind.Start(photoprism.IndexOptionsAll())
		} else {
			ind.Start(photoprism.IndexOptionsNone())
		}

		elapsed := int(time.Since(start).Seconds())

		event.Success(fmt.Sprintf("indexing completed in %d s", elapsed))
		event.Publish("index.completed", event.Data{"path": path, "seconds": elapsed})
		event.Publish("config.updated", event.Data(conf.ClientConfig()))

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("indexing completed in %d s", elapsed)})
	})
}

// DELETE /api/v1/index
func CancelIndexing(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/index", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		ind := service.Index()

		ind.Cancel()

		c.JSON(http.StatusOK, gin.H{"message": "indexing canceled"})
	})
}
