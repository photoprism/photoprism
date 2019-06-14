package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism"
)

var importer *photoprism.Importer

func initImporter(conf *config.Config) {
	if importer != nil {
		return
	}

	tensorFlow := photoprism.NewTensorFlow(conf)

	indexer := photoprism.NewIndexer(conf, tensorFlow)

	converter := photoprism.NewConverter(conf)

	importer = photoprism.NewImporter(conf, indexer, converter)
}

// POST /api/v1/import
func Import(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/import/*path", func(c *gin.Context) {
		start := time.Now()
		path := conf.ImportPath()

		if subPath := c.Param("path"); subPath != "" {
			log.Debugf("import sub path: %s", subPath)
			path = path + subPath
		}

		log.Infof("importing photos from %s", path)

		initImporter(conf)

		importer.ImportPhotosFromDirectory(path)

		elapsed := time.Since(start)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("completed import in %s", elapsed)})
	})
}
