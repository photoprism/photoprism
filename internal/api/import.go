package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// StartImport imports media files from a directory and converts/indexes them as needed.
//
// POST /api/v1/import*
func StartImport(router *gin.RouterGroup) {
	router.POST("/import/*path", func(c *gin.Context) {
		s := Auth(c, acl.ResourceFiles, acl.ActionManage)

		if s.Abort(c) {
			return
		}

		conf := service.Config()

		if conf.ReadOnly() || !conf.Settings().Features.Import {
			AbortFeatureDisabled(c)
			return
		}

		start := time.Now()

		var f form.ImportOptions

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		subPath := ""
		path := conf.ImportPath()

		if subPath = clean.Path(c.Param("path")); subPath != "" && subPath != "/" {
			subPath = strings.Replace(subPath, ".", "", -1)
			path = filepath.Join(path, subPath)
		} else if f.Path != "" {
			subPath = strings.Replace(f.Path, ".", "", -1)
			path = filepath.Join(path, subPath)
		}

		path = filepath.Clean(path)

		imp := service.Import()

		RemoveFromFolderCache(entity.RootImport)

		var opt photoprism.ImportOptions

		if f.Move {
			event.InfoMsg(i18n.MsgMovingFilesFrom, clean.Log(filepath.Base(path)))
			opt = photoprism.ImportOptionsMove(path)
		} else {
			event.InfoMsg(i18n.MsgCopyingFilesFrom, clean.Log(filepath.Base(path)))
			opt = photoprism.ImportOptionsCopy(path)
		}

		if len(f.Albums) > 0 {
			log.Debugf("import: adding files to album %s", clean.Log(strings.Join(f.Albums, " and ")))
			opt.Albums = f.Albums
		}

		imp.Start(opt)

		if subPath != "" && path != conf.ImportPath() && fs.DirIsEmpty(path) {
			if err := os.Remove(path); err != nil {
				log.Errorf("import: failed deleting empty folder %s: %s", clean.Log(path), err)
			} else {
				log.Infof("import: deleted empty folder %s", clean.Log(path))
			}
		}

		moments := service.Moments()

		if err := moments.Start(); err != nil {
			log.Warnf("moments: %s", err)
		}

		elapsed := int(time.Since(start).Seconds())

		msg := i18n.Msg(i18n.MsgImportCompletedIn, elapsed)

		event.Success(msg)
		event.Publish("import.completed", event.Data{"path": path, "seconds": elapsed})
		event.Publish("index.completed", event.Data{"path": path, "seconds": elapsed})

		for _, uid := range f.Albums {
			PublishAlbumEvent(EntityUpdated, uid, c)
		}

		UpdateClientConfig()

		// Update album, label, and subject cover thumbs.
		if err := query.UpdateCovers(); err != nil {
			log.Warnf("index: %s (update covers)", err)
		}

		c.JSON(http.StatusOK, i18n.Response{Code: http.StatusOK, Msg: msg})
	})
}

// CancelImport stops the current import operation.
//
// DELETE /api/v1/import
func CancelImport(router *gin.RouterGroup) {
	router.DELETE("/import", func(c *gin.Context) {
		s := Auth(c, acl.ResourceFiles, acl.ActionManage)

		if s.Abort(c) {
			return
		}

		conf := service.Config()

		if conf.ReadOnly() || !conf.Settings().Features.Import {
			AbortFeatureDisabled(c)
			return
		}

		imp := service.Import()

		imp.Cancel()

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgImportCanceled))
	})
}
