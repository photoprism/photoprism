package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/photoprism/photoprism/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism"
)

var indexer *photoprism.Indexer

func initIndexer(conf *config.Config) {
	if indexer != nil {
		return
	}

	tensorFlow := photoprism.NewTensorFlow(conf)

	indexer = photoprism.NewIndexer(conf, tensorFlow)
}

// POST /api/v1/index
func Index(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/index", func(c *gin.Context) {
		start := time.Now()
		path := conf.OriginalsPath()

		log.Infof("indexing photos in %s", path)

		initIndexer(conf)

		indexer.IndexAll()

		elapsed := time.Since(start)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("indexing completed in %s", elapsed)})
	})
}
