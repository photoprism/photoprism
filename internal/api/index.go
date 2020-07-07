package api

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/txt"
)

// POST /api/v1/index
func StartIndexing(router *gin.RouterGroup) {
	router.POST("/index", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if !conf.Settings().Features.Library {
			AbortFeatureDisabled(c)
			return
		}

		start := time.Now()

		var f form.IndexOptions

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		path := conf.OriginalsPath()

		ind := service.Index()

		indOpt := photoprism.IndexOptions{
			Rescan:  f.Rescan,
			Convert: f.Convert && conf.SidecarWritable(),
			Path:    filepath.Clean(f.Path),
		}

		if len(indOpt.Path) > 1 {
			event.InfoMsg(i18n.MsgIndexingFiles, txt.Quote(indOpt.Path))
		} else {
			event.InfoMsg(i18n.MsgIndexingOriginals)
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
			event.InfoMsg(i18n.MsgRemovedFilesAndPhotos, len(files), len(photos))
		}

		moments := service.Moments()

		if err := moments.Start(); err != nil {
			log.Error(err)
		}

		elapsed := int(time.Since(start).Seconds())

		msg := i18n.Msg(i18n.MsgIndexingCompletedIn, elapsed)

		event.Success(msg)
		event.Publish("index.completed", event.Data{"path": path, "seconds": elapsed})

		UpdateClientConfig()

		c.JSON(http.StatusOK, i18n.Response{Code: http.StatusOK, Msg: msg})
	})
}

// DELETE /api/v1/index
func CancelIndexing(router *gin.RouterGroup) {
	router.DELETE("/index", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if !conf.Settings().Features.Library {
			AbortFeatureDisabled(c)
			return
		}

		ind := service.Index()

		ind.Cancel()

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgIndexingCanceled))
	})
}
