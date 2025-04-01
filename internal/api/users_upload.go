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

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// UploadUserFiles adds files to the user upload folder, from where they can be moved and indexed.
//
//	@Tags	Users, Files
//	@Router /users/{uid}/upload/{token} [post]
func UploadUserFiles(router *gin.RouterGroup) {
	router.POST("/users/:uid/upload/:token", func(c *gin.Context) {
		conf := get.Config()

		// Abort in public mode or when the upload feature is disabled.
		if conf.ReadOnly() || !conf.Settings().Features.Upload {
			Abort(c, http.StatusForbidden, i18n.ErrReadOnly)
			return
		}

		// Check if the account owner is allowed to upload files.
		s := AuthAny(c, acl.ResourceFiles, acl.Permissions{acl.ActionManage, acl.ActionUpload})

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))

		// Users may only upload files for their own account.
		if s.User().UserUID != uid {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload files", "user does not match"}, s.RefID)
			AbortForbidden(c)
			return
		}

		// Abort if there is not enough free storage to upload new files.
		if conf.FilesQuotaReached() {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload files", "insufficient storage"}, s.RefID)
			Abort(c, http.StatusInsufficientStorage, i18n.ErrInsufficientStorage)
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

		var uploads []string

		// Compose upload path.
		uploadDir, err := conf.UserUploadPath(s.UserUID, s.RefID+token)

		if err != nil {
			log.Errorf("upload: failed to create storage folder (%s)", err)
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		// If the file extension list is empty, all file types may
		// be uploaded except raw files if raw support is disabled.
		allowedExt := conf.UploadAllow()
		rejectRaw := conf.DisableRaw()

		// Save uploaded files and append their names
		// to "uploads" if they pass all checks.
		for _, file := range files {
			baseName := filepath.Base(file.Filename)
			destName := path.Join(uploadDir, baseName)
			fileType := fs.FileType(baseName)

			// Reject unsupported files and files with extensions that aren't allowed.
			if fileType == fs.TypeUnknown {
				log.Warnf("upload: rejected %s because it has an unsupported file extension", clean.Log(baseName))
				continue
			} else if allowedExt.Excludes(fileType.DefaultExt()) {
				log.Warnf("upload: rejected %s because its extension is not allowed", clean.Log(baseName))
				continue
			}

			// Save uploaded file in the user upload path.
			if err = c.SaveUploadedFile(file, destName); err != nil {
				log.Errorf("upload: failed to save %s", clean.Log(baseName))
				log.Debugf("upload: %s in %s", clean.Error(err), clean.Log(baseName))
				Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
				return
			} else {
				log.Debugf("upload: saved %s in user upload path", clean.Log(baseName))
				event.Publish("upload.saved", event.Data{"uid": s.UserUID, "file": baseName})
			}

			// Make sure the file is supported and has the correct extension before importing it.
			if mediaFile, mediaErr := photoprism.NewMediaFile(destName); mediaErr != nil {
				log.Errorf("upload: rejected %s, %s", clean.Error(err), clean.Log(baseName))
				logErr("upload", os.Remove(destName))
			} else if typeErr := mediaFile.CheckType(); typeErr != nil {
				log.Warnf("upload: rejected %s %s", clean.Log(baseName), typeErr)
				logErr("upload", os.Remove(destName))
			} else if rejectRaw && mediaFile.IsRaw() {
				log.Warnf("upload: rejected %s because raw support is disabled", clean.Log(baseName))
				logErr("upload", os.Remove(destName))
			} else {
				// Successfully validated upload.
				uploads = append(uploads, destName)
			}
		}

		// Check if the uploaded file may contain inappropriate content.
		if len(uploads) > 0 && !conf.UploadNSFW() {
			nd := get.NsfwDetector()

			containsNSFW := false

			for _, filename := range uploads {
				labels, nsfwErr := nd.File(filename)

				if nsfwErr != nil {
					log.Debug(nsfwErr)
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

		// Log number of successfully uploaded files.
		msg := i18n.Msg(i18n.MsgFilesUploadedIn, len(uploads), elapsed)

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

		var frm form.UploadOptions

		// Assign and validate request form values.
		if err := c.BindJSON(&frm); err != nil {
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
		if len(frm.Albums) > 0 &&
			acl.Rules.AllowAny(acl.ResourceAlbums, s.UserRole(), acl.Permissions{acl.ActionCreate, acl.ActionUpload}) {
			log.Debugf("upload: adding files to album %s", clean.Log(strings.Join(frm.Albums, " and ")))
			opt.Albums = frm.Albums
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
		if n := imported.Processed(); n == 0 {
			log.Infof("upload: found no new files to import from %s", clean.Log(uploadPath))
		} else {
			log.Infof("upload: imported %s", english.Plural(n, "file", "files"))
			if moments := get.Moments(); moments == nil {
				log.Warnf("upload: moments service not set - you may have found a bug")
			} else if workerErr := moments.Start(); workerErr != nil {
				log.Warnf("moments: %s", workerErr)
			}
		}

		elapsed := int(time.Since(start).Seconds())

		// Show success message.
		msg := i18n.Msg(i18n.MsgUploadProcessed)

		event.Success(msg)
		event.Publish("import.completed", event.Data{"uid": opt.UID, "path": uploadPath, "seconds": elapsed})
		event.Publish("index.completed", event.Data{"uid": opt.UID, "path": uploadPath, "seconds": elapsed})
		event.Publish("upload.completed", event.Data{"uid": opt.UID, "path": uploadPath, "seconds": elapsed})

		for _, uid := range frm.Albums {
			PublishAlbumEvent(StatusUpdated, uid, c)
		}

		// Update the user interface.
		UpdateClientConfig()

		// Update album, label, and subject cover thumbs.
		if coversErr := query.UpdateCovers(); coversErr != nil {
			log.Warnf("upload: %s (update covers)", coversErr)
		}

		c.JSON(http.StatusOK, i18n.Response{Code: http.StatusOK, Msg: msg})
	})
}
