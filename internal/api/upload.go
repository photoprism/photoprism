package api

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/gin-gonic/gin"
)

// POST /api/v1/upload/:path
func Upload(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/upload/:path", func(c *gin.Context) {
		if conf.ReadOnly() {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrReadOnly)
			return
		}

		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		start := time.Now()
		subPath := c.Param("path")

		f, err := c.MultipartForm()

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		event.Publish("upload.start", event.Data{"time": start})

		files := f.File["files"]
		uploaded := len(files)
		var uploads []string

		p := path.Join(conf.ImportPath(), "upload", subPath)

		if err := os.MkdirAll(p, os.ModePerm); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		for _, file := range files {
			filename := path.Join(p, filepath.Base(file.Filename))

			log.Debugf("upload: saving file \"%s\"", file.Filename)

			if err := c.SaveUploadedFile(file, filename); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
				return
			}

			uploads = append(uploads, filename)
		}

		if !conf.UploadNSFW() {
			initNsfwDetector(conf)

			containsNSFW := false

			for _, filename := range uploads {
				labels, err := nd.File(filename)

				if err != nil {
					log.Debug(err)
					continue
				}

				if labels.IsSafe() {
					continue
				}

				log.Infof("nsfw: \"%s\" might be offensive", filename)

				containsNSFW = true
			}

			if containsNSFW {
				for _, filename := range uploads {
					if err := os.Remove(filename); err != nil {
						log.Errorf("nsfw: could not delete \"%s\"", filename)
					}
				}

				c.AbortWithStatusJSON(http.StatusForbidden, ErrUploadNSFW)
				return
			}
		}

		elapsed := time.Since(start)

		log.Infof("%d files uploaded in %s", uploaded, elapsed)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d files uploaded in %s", uploaded, elapsed)})
	})
}
