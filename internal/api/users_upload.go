package api

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/txt"
)

// UploadUserFiles adds files to the user's upload folder from where they can be processed and indexed.
//
//	@Summary	upload files to a user's upload folder
//	@Id			UploadUserFiles
//	@Tags		Users, Files
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		uid						path		string	true	"user uid"
//	@Param		token					path		string	true	"upload token"
//	@Param		files					formData	file	true	"one or more files to upload (repeat the field for multiple files)"
//	@Success	200						{object}	i18n.Response
//	@Failure	400,401,403,413,429,507	{object}	i18n.Response
//	@Router		/api/v1/users/{uid}/upload/{token} [post]
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
		if s.GetUser().UserUID != uid {
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
		rejectArchives := !conf.UploadArchives()
		rejectRaw := conf.DisableRaw()
		fileSizeLimit := conf.OriginalsLimitBytes()
		totalSizeLimit := conf.UploadLimitBytes()

		// Save uploaded files and append their names
		// to "uploads" if they pass all checks.
		for _, file := range files {
			baseName := filepath.Base(file.Filename)
			destName := path.Join(uploadDir, baseName)
			fileType := fs.FileType(baseName)

			// Reject unsupported files and files with extensions that aren't allowed.
			if fileType == fs.TypeUnknown {
				log.Errorf("upload: rejected %s because it has an unsupported file extension", clean.Log(baseName))
				continue
			} else if allowedExt.Excludes(fileType.DefaultExt()) {
				log.Errorf("upload: rejected %s because its extension is not allowed", clean.Log(baseName))
				continue
			} else if fileSizeLimit > 0 && file.Size > fileSizeLimit {
				log.Errorf("upload: rejected %s because its size exceeds the file size limit", clean.Log(baseName))
				continue
			}

			// Save uploaded file in the user upload path.
			if err = c.SaveUploadedFile(file, destName); err != nil {
				log.Debugf("upload: %s in %s", clean.Error(err), clean.Log(baseName))
				log.Errorf("upload: failed to save %s", clean.Log(baseName))
				Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
				return
			} else {
				log.Debugf("upload: saved %s in user upload path", clean.Log(baseName))
				event.Publish("upload.saved", event.Data{"uid": s.UserUID, "file": baseName})
			}

			// Extract contents if the uploaded file is an archive.
			if ext := fs.ArchiveExt(baseName); ext != "" {
				if rejectArchives {
					logWarn("upload", os.Remove(destName))
					log.Errorf("upload: rejected %s because archive uploads are disabled", clean.Log(baseName))
					continue
				}

				zipFiles, skippedFiles, zipErr := fs.Unzip(destName, uploadDir, fileSizeLimit, totalSizeLimit)

				logWarn("upload", os.Remove(destName))

				if zipErr != nil {
					log.Errorf("upload: failed to extract files from %s (%s)", clean.Log(baseName), zipErr)
				}

				if len(skippedFiles) > 0 {
					log.Errorf("upload: could not extract %s from %s due to upload restrictions", strings.Join(skippedFiles, ", "), clean.Log(baseName))
				}

				if len(zipFiles) == 0 {
					continue
				}

				for _, destName = range zipFiles {
					baseName = filepath.Base(destName)
					fileType = fs.FileType(baseName)

					// Reject unsupported files and files with extensions that aren't allowed.
					if baseName == "" {
						log.Errorf("upload: rejected unzipped file because it has no file name")
					} else if baseName[0] == '.' || baseName[0] == '@' {
						logWarn("upload", os.Remove(destName))
						log.Errorf("upload: rejected unzipped file %s because it has an unsupported file name", clean.Log(baseName))
					} else if fileType == fs.TypeUnknown {
						logWarn("upload", os.Remove(destName))
						log.Errorf("upload: rejected unzipped file %s because it has an unsupported file extension", clean.Log(baseName))
					} else if allowedExt.Excludes(fileType.DefaultExt()) {
						logWarn("upload", os.Remove(destName))
						log.Errorf("upload: rejected unzipped file %s because its extension is not allowed", clean.Log(baseName))
					} else if totalSizeLimit, err = UploadCheckFile(destName, rejectRaw, totalSizeLimit); err != nil {
						log.Errorf("upload: %s", err)
					} else {
						// Add to the list of uploaded files after having verified that
						// the unzipped file has the correct extension and format.
						uploads = append(uploads, destName)
					}
				}
			} else if totalSizeLimit, err = UploadCheckFile(destName, rejectRaw, totalSizeLimit); err != nil {
				log.Errorf("upload: %s", err)
			} else {
				// Add to the list of uploaded files after having verified that
				// the uploaded file has the correct extension and format.
				uploads = append(uploads, destName)
			}
		}

		// Check if the uploaded file may contain inappropriate content.
		if len(uploads) > 0 && !conf.UploadNSFW() {
			containsNSFW := false

			for _, filename := range uploads {
				labels, nsfwErr := vision.DetectNSFW([]string{filename}, media.SrcLocal)

				if nsfwErr != nil {
					log.Debug(nsfwErr)
					continue
				} else if len(labels) < 1 {
					log.Errorf("nsfw: model returned no result")
					continue
				} else if labels[0].IsSafe() {
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

// UploadCheckFile checks if the file is supported and has the correct extension.
func UploadCheckFile(destName string, rejectRaw bool, totalSizeLimit int64) (remainingSizeLimit int64, err error) {
	baseName := filepath.Base(destName)

	if mediaFile, mediaErr := photoprism.NewMediaFile(destName); mediaErr != nil {
		logWarn("upload", os.Remove(destName))
		return totalSizeLimit, fmt.Errorf("rejected %s, %s", clean.Error(err), clean.Log(baseName))
	} else if typeErr := mediaFile.CheckType(); typeErr != nil {
		logWarn("upload", os.Remove(destName))
		return totalSizeLimit, fmt.Errorf("rejected %s %s", clean.Log(baseName), typeErr)
	} else if rejectRaw && mediaFile.IsRaw() {
		logWarn("upload", os.Remove(destName))
		return totalSizeLimit, fmt.Errorf("rejected %s because raw support is disabled", clean.Log(baseName))
	} else if totalSizeLimit < 0 {
		return -1, nil
	} else if remainingSizeLimit = totalSizeLimit - mediaFile.FileSize(); totalSizeLimit == 0 || remainingSizeLimit < 1 {
		logWarn("upload", os.Remove(destName))
		return 0, fmt.Errorf("rejected %s because the total upload size limit has been reached", clean.Log(baseName))
	} else {
		return remainingSizeLimit, nil
	}
}

// ProcessUserUpload triggers processing and import of previously uploaded files.
//
//	@Summary	process previously uploaded files for a user
//	@Id			ProcessUserUpload
//	@Tags		Users, Files
//	@Accept		json
//	@Produce	json
//	@Param		uid						path		string				true	"user uid"
//	@Param		token					path		string				true	"upload token"
//	@Param		options					body		form.UploadOptions	true	"processing options"
//	@Success	200						{object}	i18n.Response
//	@Failure	400,401,403,404,409,429	{object}	i18n.Response
//	@Router		/api/v1/users/{uid}/upload/{token} [put]
func ProcessUserUpload(router *gin.RouterGroup) {
	router.PUT("/users/:uid/upload/:token", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourceFiles, acl.Permissions{acl.ActionManage, acl.ActionUpload})

		if s.Abort(c) {
			return
		}

		// Users may only upload their own files.
		if s.GetUser().UserUID != clean.UID(c.Param("uid")) {
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
			AbortBadRequest(c, err)
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
		if destFolder = s.GetUser().GetUploadPath(); destFolder == "" {
			destFolder = conf.ImportDest()
		}

		// Move uploaded files to the destination folder.
		event.InfoMsg(i18n.MsgProcessingUpload)
		opt := photoprism.ImportOptionsUpload(uploadPath, destFolder)

		// Add imported files to albums if allowed.
		if len(frm.Albums) > 0 &&
			acl.Rules.AllowAny(acl.ResourceAlbums, s.GetUserRole(), acl.Permissions{acl.ActionCreate, acl.ActionUpload}) {
			log.Debugf("upload: adding files to album %s", clean.Log(txt.JoinAnd(frm.Albums)))
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
