package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/fs"
)

var imp *photoprism.Import

func initImport(conf *config.Config) {
	if imp != nil {
		return
	}

	initIndex(conf)

	convert := photoprism.NewConvert(conf)

	imp = photoprism.NewImport(conf, ind, convert)
}

// POST /api/v1/import*
func StartImport(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/import/*path", func(c *gin.Context) {
		if conf.ReadOnly() {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrReadOnly)
			return
		}

		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		subPath := ""
		start := time.Now()
		path := conf.ImportPath()

		if subPath = c.Param("path"); subPath != "" && subPath != "/" {
			subPath = strings.Replace(subPath, ".", "", -1)
			log.Debugf("import sub path: %s", subPath)
			path = path + subPath
		}

		path = filepath.Clean(path)

		event.Info(fmt.Sprintf("importing photos from \"%s\"", filepath.Base(path)))

		initImport(conf)

		imp.Start(path)

		if subPath != "" && path != conf.ImportPath() && fs.IsEmpty(path) {
			if err := os.Remove(path); err != nil {
				log.Errorf("import: could not deleted empty directory \"%s\": %s", path, err)
			} else {
				log.Infof("import: deleted empty directory \"%s\"", path)
			}
		}

		elapsed := int(time.Since(start).Seconds())

		event.Success(fmt.Sprintf("import completed in %d s", elapsed))
		event.Publish("import.completed", event.Data{"path": path, "seconds": elapsed})
		event.Publish("index.completed", event.Data{"path": path, "seconds": elapsed})
		event.Publish("config.updated", event.Data(conf.ClientConfig()))

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("import completed in %d s", elapsed)})
	})
}

// DELETE /api/v1/import
func CancelImport(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/import", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		initImport(conf)

		imp.Cancel()

		c.JSON(http.StatusOK, gin.H{"message": "import canceled"})
	})
}
