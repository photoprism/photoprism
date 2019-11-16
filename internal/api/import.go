package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism"
)

var importer *photoprism.Importer

func initImporter(conf *config.Config) {
	if importer != nil {
		return
	}

	initIndexer(conf)

	converter := photoprism.NewConverter(conf)

	importer = photoprism.NewImporter(conf, indexer, converter)
}

// POST /api/v1/import
func Import(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/import/*path", func(c *gin.Context) {
		if conf.ReadOnly() {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrReadOnly)
			return
		}

		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		start := time.Now()
		path := conf.ImportPath()

		if subPath := c.Param("path"); subPath != "" {
			log.Debugf("import sub path: %s", subPath)
			path = path + subPath
		}

		event.Info(fmt.Sprintf("importing photos from \"%s\"", filepath.Base(path)))

		initImporter(conf)

		importer.ImportPhotosFromDirectory(path)

		elapsed := int(time.Since(start).Seconds())

		event.Success(fmt.Sprintf("import completed in %d s", elapsed))
		event.Publish("import.completed", event.Data{"path": path, "seconds": elapsed})
		event.Publish("index.completed", event.Data{"path": path, "seconds": elapsed})
		event.Publish("config.updated", event.Data(conf.ClientConfig()))

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("import completed in %d s", elapsed)})
	})
}
