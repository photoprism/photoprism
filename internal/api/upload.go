package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"
	uuid "github.com/satori/go.uuid"
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

// POST /api/v1/upload
func Upload(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/upload", func(c *gin.Context) {
		start := time.Now()

		form, err := c.MultipartForm()

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		log.Debugf("Value: %#v", form.Value)
		log.Debugf("File: %#v", form.File)

		files := form.File["files"]

		path := fmt.Sprintf("%s/uploads/%s", conf.ImportPath(), uuid.NewV4())

		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		for _, file := range files {
			filename := fmt.Sprintf("%s/%s", path, filepath.Base(file.Filename))

			if err := c.SaveUploadedFile(file, filename); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
				return
			}
		}

		log.Infof("importing photos from %s", conf.ImportPath())

		initImporter(conf)

		importer.ImportPhotosFromDirectory(conf.ImportPath())

		elapsed := time.Since(start)

		log.Infof("%d files imported in %s", len(files), elapsed)

		c.JSON(http.StatusOK,  gin.H{"message": fmt.Sprintf("%d files imported in %s", len(files), elapsed)})
	})
}
