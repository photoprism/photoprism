package api

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
)

// POST /api/v1/upload/:path
func Upload(router *gin.RouterGroup) {
	router.POST("/upload/:path", func(c *gin.Context) {
		conf := service.Config()
		if conf.ReadOnly() || !conf.Settings().Features.Upload {
			Abort(c, http.StatusForbidden, i18n.ErrReadOnly)
			return
		}

		s := Auth(c, acl.ResourcePhotos, acl.ActionUpload)

		if s.Abort(c) {
			return
		}

		start := time.Now()
		subPath := clean.Path(c.Param("path"))

		f, err := c.MultipartForm()

		if err != nil {
			log.Errorf("upload: %s", err)
			AbortBadRequest(c)
			return
		}

		event.Publish("upload.start", event.Data{"time": start})

		files := f.File["files"]
		uploaded := len(files)
		var uploads []string

		p := path.Join(conf.ImportPath(), "upload", subPath)

		if err := os.MkdirAll(p, os.ModePerm); err != nil {
			log.Errorf("upload: failed creating folder %s", clean.Log(subPath))
			AbortBadRequest(c)
			return
		}

		for _, file := range files {
			filename := path.Join(p, filepath.Base(file.Filename))

			log.Debugf("upload: saving file %s", clean.Log(file.Filename))

			if err := c.SaveUploadedFile(file, filename); err != nil {
				log.Errorf("upload: failed saving file %s", clean.Log(filepath.Base(file.Filename)))
				AbortBadRequest(c)
				return
			}

			uploads = append(uploads, filename)
		}

		if !conf.UploadNSFW() {
			nd := service.NsfwDetector()

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

				log.Infof("nsfw: %s might be offensive", clean.Log(filename))

				containsNSFW = true
			}

			if containsNSFW {
				for _, filename := range uploads {
					if err := os.Remove(filename); err != nil {
						log.Errorf("nsfw: could not delete %s", clean.Log(filename))
					}
				}

				Abort(c, http.StatusForbidden, i18n.ErrOffensiveUpload)
				return
			}
		}

		elapsed := int(time.Since(start).Seconds())

		msg := i18n.Msg(i18n.MsgFilesUploadedIn, uploaded, elapsed)

		log.Info(msg)

		c.JSON(http.StatusOK, i18n.Response{Code: http.StatusOK, Msg: msg})
	})
}
