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
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// POST /api/v1/import*
func StartImport(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/import/*path", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		if conf.ReadOnly() || !conf.Settings().Features.Import {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrFeatureDisabled)
			return
		}

		start := time.Now()

		var f form.ImportOptions

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		subPath := ""
		path := conf.ImportPath()

		if subPath = c.Param("path"); subPath != "" && subPath != "/" {
			subPath = strings.Replace(subPath, ".", "", -1)
			path = filepath.Join(path, subPath)
		} else if f.Path != "" {
			subPath = strings.Replace(f.Path, ".", "", -1)
			path = filepath.Join(path, subPath)
		}

		path = filepath.Clean(path)

		imp := service.Import()

		var opt photoprism.ImportOptions

		if f.Move {
			event.Info(fmt.Sprintf("moving files from %s", txt.Quote(filepath.Base(path))))
			opt = photoprism.ImportOptionsMove(path)
		} else {
			event.Info(fmt.Sprintf("copying files from %s", txt.Quote(filepath.Base(path))))
			opt = photoprism.ImportOptionsCopy(path)
		}

		if len(f.Albums) > 0 {
			log.Debugf("import: files will be added to album %s", strings.Join(f.Albums, " and "))
			opt.Albums = f.Albums
		}

		imp.Start(opt)

		if subPath != "" && path != conf.ImportPath() && fs.IsEmpty(path) {
			if err := os.Remove(path); err != nil {
				log.Errorf("import: could not delete empty folder %s: %s", txt.Quote(path), err)
			} else {
				log.Infof("import: deleted empty folder %s", txt.Quote(path))
			}
		}

		moments := service.Moments()

		if err := moments.Start(); err != nil {
			log.Error(err)
		}

		elapsed := int(time.Since(start).Seconds())

		event.Success(fmt.Sprintf("import completed in %d s", elapsed))
		event.Publish("import.completed", event.Data{"path": path, "seconds": elapsed})
		event.Publish("index.completed", event.Data{"path": path, "seconds": elapsed})

		for _, uid := range f.Albums {
			PublishAlbumEvent(EntityUpdated, uid, c)
		}

		UpdateClientConfig(conf)

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

		imp := service.Import()

		imp.Cancel()

		c.JSON(http.StatusOK, gin.H{"message": "import canceled"})
	})
}
