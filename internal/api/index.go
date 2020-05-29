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

		ind := service.Index()

		indOpt := photoprism.IndexOptions{
			Rescan:  f.Rescan,
			Convert: f.Convert && !conf.ReadOnly(),
			Path:    filepath.Clean(f.Path),
		}

		if len(indOpt.Path) > 1 {
			event.Info(fmt.Sprintf("indexing files in %s", txt.Quote(indOpt.Path)))
		} else {
			event.Info("indexing originals...")
		}

		indexed := ind.Start(indOpt)

		prg := service.Purge()

		prgOpt := photoprism.PurgeOptions{
			Path:   filepath.Clean(f.Path),
			Ignore: indexed,
		}

		if files, photos, err := prg.Start(prgOpt); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst(err.Error())})
			return
		} else if len(files) > 0 || len(photos) > 0 {
			event.Info(fmt.Sprintf("removed %d files and %d photos", len(files), len(photos)))
		}

		moments := service.Moments()

		if err := moments.Start(); err != nil {
			log.Error(err)
		}

		elapsed := int(time.Since(start).Seconds())

		event.Success(fmt.Sprintf("indexing completed in %d s", elapsed))
		event.Publish("index.completed", event.Data{"path": path, "seconds": elapsed})

		UpdateClientConfig(conf)

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
