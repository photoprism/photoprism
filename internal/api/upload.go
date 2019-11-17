package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"

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

		form, err := c.MultipartForm()

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		files := form.File["files"]

		path := fmt.Sprintf("%s/upload/%s", conf.ImportPath(), subPath)

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

		elapsed := time.Since(start)

		log.Infof("%d files uploaded in %s", len(files), elapsed)

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d files uploaded in %s", len(files), elapsed)})
	})
}
