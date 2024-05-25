package api

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// UploadUserFiles adds files to the user upload folder, from where they can be moved and indexed.
//
// POST /users/:uid/upload/:token
func UploadUserFiles(router *gin.RouterGroup) {
	router.POST("/users/:uid/upload/:token", func(c *gin.Context) {
		conf := get.Config()

		// Abort in public mode or when the upload feature is disabled.
		if conf.ReadOnly() || !conf.Settings().Features.Upload {
			Abort(c, http.StatusForbidden, i18n.ErrReadOnly)
			return
		}

		// Check permission.
		s := AuthAny(c, acl.ResourceFiles, acl.Permissions{acl.ActionManage, acl.ActionUpload})

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))

		// Users may only upload their own files.
		if s.User().UserUID != uid {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload files", "user does not match"}, s.RefID)
			AbortForbidden(c)
			return
		}

		start := time.Now()
		token := clean.Token(c.Param("token"))

		f, err := c.MultipartForm()

		if err != nil {
			log.Errorf("upload: %s", err)
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		// Publish upload start event.
		event.Publish("upload.start", event.Data{"uid": s.UserUID, "time": start})

		files := f.File["files"]
		uploaded := len(files)

		var uploads []string

		// Compose upload path.
		uploadDir, err := conf.UserUploadPath(s.UserUID, s.RefID+token)

		if err != nil {
			log.Errorf("upload: failed to create storage folder (%s)", err)
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		// Save uploaded files.
		for _, file := range files {
			fileName := filepath.Base(file.Filename)
			filePath := path.Join(uploadDir, fileName)

			if err = c.SaveUploadedFile(file, filePath); err != nil {
				log.Errorf("upload: failed saving file %s", clean.Log(fileName))
				Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
				return
			} else {
				log.Debugf("upload: saved file %s", clean.Log(fileName))
				event.Publish("upload.saved", event.Data{"uid": s.UserUID, "file": fileName})
			}

			uploads = append(uploads, filePath)
		}

		// Check if uploaded file is safe.
		if !conf.UploadNSFW() {
			nd := get.NsfwDetector()

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

// ProcessUserUpload triggers processing once all files have been uploaded.
//
// PUT /users/:uid/upload/:token
func ProcessUserUpload(router *gin.RouterGroup) {
	router.PUT("/users/:uid/upload/:token", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourceFiles, acl.Permissions{acl.ActionManage, acl.ActionUpload})

		if s.Abort(c) {
			return
		}

		// Users may only upload their own files.
		if s.User().UserUID != clean.UID(c.Param("uid")) {
			AbortForbidden(c)
			return
		}

		conf := get.Config()

		if conf.ReadOnly() || !conf.Settings().Features.Import {
			AbortFeatureDisabled(c)
			return
		}

		start := time.Now()

		var f form.UploadOptions

		// Assign and validate request form values.
		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		token := clean.Token(c.Param("token"))
		uploadPath, err := conf.UserUploadPath(s.UserUID, s.RefID+token)

		if err != nil {
			log.Errorf("upload: failed to create storage folder (%s)", err)
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		imp := get.Import()

		// Get destination folder.
		var destFolder string
		if destFolder = s.User().GetUploadPath(); destFolder == "" {
			destFolder = conf.ImportDest()
		}

		// Move uploaded files to the destination folder.
		event.InfoMsg(i18n.MsgProcessingUpload)
		opt := photoprism.ImportOptionsUpload(uploadPath, destFolder)

		// Add imported files to albums if allowed.
		if len(f.Albums) > 0 &&
			acl.Rules.AllowAny(acl.ResourceAlbums, s.UserRole(), acl.Permissions{acl.ActionCreate, acl.ActionUpload}) {
			log.Debugf("upload: adding files to album %s", clean.Log(strings.Join(f.Albums, " and ")))
			opt.Albums = f.Albums
		}

		// Set user UID if known.
		if s.UserUID != "" {
			opt.UID = s.UserUID
		}

		// Start import.
		imported := imp.Start(opt)

		// Delete empty import directory.
		if fs.DirIsEmpty(uploadPath) {
			if err := os.Remove(uploadPath); err != nil {
				log.Errorf("upload: failed to delete empty folder %s: %s", clean.Log(uploadPath), err)
			} else {
				log.Infof("upload: deleted empty folder %s", clean.Log(uploadPath))
			}
		}

		// Update moments if files have been imported.
		if n := len(imported); n == 0 {
			log.Infof("upload: found no new files to import from %s", clean.Log(uploadPath))
		} else {
			log.Infof("upload: imported %s", english.Plural(n, "file", "files"))
			if moments := get.Moments(); moments == nil {
				log.Warnf("upload: moments service not set - you may have found a bug")
			} else if err := moments.Start(); err != nil {
				log.Warnf("moments: %s", err)
			}
		}

		elapsed := int(time.Since(start).Seconds())

		// Show success message.
		msg := i18n.Msg(i18n.MsgUploadProcessed)

		event.Success(msg)
		event.Publish("import.completed", event.Data{"uid": opt.UID, "path": uploadPath, "seconds": elapsed})
		event.Publish("index.completed", event.Data{"uid": opt.UID, "path": uploadPath, "seconds": elapsed})
		event.Publish("upload.completed", event.Data{"uid": opt.UID, "path": uploadPath, "seconds": elapsed})

		for _, uid := range f.Albums {
			PublishAlbumEvent(StatusUpdated, uid, c)
		}

		// Update the user interface.
		UpdateClientConfig()

		// Update album, label, and subject cover thumbs.
		if err := query.UpdateCovers(); err != nil {
			log.Warnf("upload: %s (update covers)", err)
		}

		c.JSON(http.StatusOK, i18n.Response{Code: http.StatusOK, Msg: msg})
	})
}
