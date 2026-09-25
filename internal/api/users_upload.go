package api

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
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

		// Users whose access is limited to their base path need an upload path.
		if uploadPathDenied(s.GetUser()) {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload files", "no upload path", status.Denied}, s.RefID)
			AbortForbidden(c)
			return
		}

		// Abort if there is not enough free storage to upload new files.
		if conf.InsufficientStorage() {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload files", status.InsufficientStorage}, s.RefID)
			Abort(c, http.StatusInsufficientStorage, i18n.ErrInsufficientStorage)
			return
		}

		start := time.Now()
		token := clean.Token(c.Param("token"))

		if totalSizeLimit := conf.UploadLimitBytes(); totalSizeLimit > 0 {
			LimitRequestBodyBytes(c, totalSizeLimit+MaxMultipartOverheadBytes)
		}

		f, err := c.MultipartForm()

		if err != nil {
			if IsRequestBodyTooLarge(err) {
				log.Errorf("upload: %s", clean.Error(err))
				AbortRequestTooLarge(c, i18n.ErrFileTooLarge)
				return
			}

			log.Errorf("upload: %s", clean.Error(err))
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
			log.Errorf("upload: failed to create storage folder (%s)", clean.Error(err))
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		// Operator extension settings can further restrict the supported upload formats.
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
			switch {
			case fileType == fs.TypeUnknown:
				log.Errorf("upload: rejected %s because it has an unsupported file extension", clean.Log(baseName))
				continue
			case !uploadSidecarAllowed(baseName):
				log.Errorf("upload: rejected %s because its sidecar format is not supported", clean.Log(baseName))
				continue
			case allowedExt.Excludes(fileType.DefaultExt()):
				log.Errorf("upload: rejected %s because its extension is not allowed", clean.Log(baseName))
				continue
			case fileSizeLimit > 0 && file.Size > fileSizeLimit:
				log.Errorf("upload: rejected %s because its size exceeds the file size limit", clean.Log(baseName))
				continue
			}

			// Save uploaded file in the user upload path.
			if err = c.SaveUploadedFile(file, destName); err != nil {
				log.Debugf("upload: %s in %s", clean.Error(err), clean.Log(baseName))

				// Report a disk-full write failure as insufficient storage so the cause is clear.
				if disk.IsNoSpace(err) {
					disk.FlushFree()
					event.AuditErr([]string{ClientIP(c), "session %s", "upload files", status.InsufficientStorage}, s.RefID)
					Abort(c, http.StatusInsufficientStorage, i18n.ErrInsufficientStorage)
					return
				}

				log.Errorf("upload: failed to save %s", clean.Log(baseName))
				Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
				return
			} else {
				log.Debugf("upload: saved %s in user upload path", clean.Log(baseName))
				event.Publish("upload.saved", event.Data{"uid": s.UserUID})
			}

			// Extract contents if the uploaded file is an archive.
			if ext := fs.ArchiveExt(baseName); ext != "" {
				if rejectArchives {
					logWarn("upload", os.Remove(destName))
					log.Errorf("upload: rejected %s because archive uploads are disabled", clean.Log(baseName))
					continue
				}

				zipFiles, skippedFiles, zipErr := fs.Unzip(destName, uploadDir, fileSizeLimit, totalSizeLimit, uploadArchiveEntryAllowed)

				logWarn("upload", os.Remove(destName))

				if zipErr != nil {
					log.Errorf("upload: failed to extract files from %s (%s)", clean.Log(baseName), clean.Error(zipErr))
				}

				if len(skippedFiles) > 0 {
					log.Errorf("upload: could not extract %d entries from %s due to upload restrictions (%s)", len(skippedFiles), clean.Log(baseName), clean.LogNames(skippedFiles))
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
						log.Errorf("upload: %s", clean.Error(err))
					} else {
						// Add to the list of uploaded files after having verified that
						// the unzipped file has the correct extension and format.
						uploads = append(uploads, destName)
					}
				}
			} else if totalSizeLimit, err = UploadCheckFile(destName, rejectRaw, totalSizeLimit); err != nil {
				log.Errorf("upload: %s", clean.Error(err))
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

				switch {
				case nsfwErr != nil:
					log.Debugf("nsfw: %s", clean.Error(nsfwErr))
					continue
				case len(labels) < 1:
					log.Errorf("nsfw: model returned no result")
					continue
				case labels[0].IsSafe():
					continue
				}

				log.Infof("nsfw: %s might be offensive", clean.Log(filepath.Base(filename)))

				containsNSFW = true
			}

			if containsNSFW {
				for _, filename := range uploads {
					if err := os.Remove(filename); err != nil {
						log.Errorf("nsfw: could not delete %s", clean.Log(filepath.Base(filename)))
					}
				}

				Abort(c, http.StatusForbidden, i18n.ErrOffensiveUpload)
				return
			}
		}

		elapsed := time.Since(start)

		// Log number of successfully uploaded files.
		log.Infof("library: uploaded %s in %s", english.Plural(len(uploads), "file", "files"), elapsed)

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgFilesUploadedIn, len(uploads), int(elapsed.Seconds())))
	})
}

// uploadPathDenied reports whether the user may not upload files because their access is limited to
// their base path and they have no upload path.
func uploadPathDenied(u *entity.User) bool {
	return u.RequiresBasePath() && u.GetUploadPath() == ""
}

// MaxUploadAlbums is the number of distinct albums an upload or import may add its files to.
const MaxUploadAlbums = 100

// tooManyUploadAlbums reports whether the list names more distinct albums than MaxUploadAlbums.
func tooManyUploadAlbums(albums []string) bool {
	if len(albums) <= MaxUploadAlbums {
		return false
	}

	seen := make(map[string]struct{}, MaxUploadAlbums+1)

	for _, album := range albums {
		if album == "" {
			continue
		}

		seen[album] = struct{}{}

		if len(seen) > MaxUploadAlbums {
			return true
		}
	}

	return false
}

// uploadAlbumsAllowed reports whether the session may add the files it uploads or imports to albums, which
// requires a user account, and permission and scope to create or upload to albums.
func uploadAlbumsAllowed(s *entity.Session) bool {
	perms := acl.Permissions{acl.ActionCreate, acl.ActionUpload}

	return s != nil && s.GetUser().IsRegistered() && s.GrantsAny(acl.ResourceAlbums, perms) && s.ValidateScope(acl.ResourceAlbums, perms)
}

// uploadAlbums returns the albums that files the session uploads or imports may be added to, without
// duplicates: titles, which resolve among the user's own albums or create a new one, and the UIDs of
// albums the session can see.
func uploadAlbums(c *gin.Context, s *entity.Session, albums []string) []string {
	result := make([]string, 0, len(albums))
	seen := make(map[string]struct{}, len(albums))
	denied := 0

	for _, album := range albums {
		if _, ok := seen[album]; ok {
			continue
		}

		seen[album] = struct{}{}

		if !rnd.IsUID(album, entity.AlbumUID) {
			result = append(result, album)
		} else if found, err := query.AlbumByUID(album); err == nil && found.HasID() && !found.Deleted() && found.VisibleToSession(s) {
			result = append(result, album)
		} else {
			denied++
		}
	}

	if denied > 0 {
		event.AuditWarn([]string{ClientIP(c), "session %s", "add files to %s", status.Denied}, s.RefID, english.Plural(denied, "album", "albums"))
	}

	return result
}

// UploadCheckFile checks if the file is supported and has the correct extension.
func UploadCheckFile(destName string, rejectRaw bool, totalSizeLimit int64) (remainingSizeLimit int64, err error) {
	baseName := filepath.Base(destName)

	if fs.FileType(baseName) != fs.TypeUnknown && !uploadSidecarAllowed(baseName) {
		logWarn("upload", os.Remove(destName))
		return totalSizeLimit, fmt.Errorf("rejected %s because its sidecar format is not supported", clean.Log(baseName))
	}

	if mediaFile, mediaErr := photoprism.NewMediaFile(destName); mediaErr != nil {
		logWarn("upload", os.Remove(destName))
		return totalSizeLimit, fmt.Errorf("rejected %s (%w)", clean.Log(baseName), mediaErr)
	} else if typeErr := mediaFile.CheckType(); typeErr != nil {
		logWarn("upload", os.Remove(destName))
		return totalSizeLimit, fmt.Errorf("rejected %s (%w)", clean.Log(baseName), typeErr)
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

		// Users whose access is limited to their base path need an upload path.
		if uploadPathDenied(s.GetUser()) {
			event.AuditErr([]string{ClientIP(c), "session %s", "import uploads", "no upload path", status.Denied}, s.RefID)
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
		LimitRequestBodyBytes(c, MaxUploadOptionsRequestBytes)

		if err := c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}

			AbortBadRequest(c, err)
			return
		}

		// Refuse to add the files to more albums than an upload may name.
		if tooManyUploadAlbums(frm.Albums) {
			event.AuditWarn([]string{ClientIP(c), "session %s", "add files to more than %d albums", status.Denied}, s.RefID, MaxUploadAlbums)
			Abort(c, http.StatusBadRequest, i18n.ErrBadRequest)
			return
		}

		token := clean.Token(c.Param("token"))
		uploadPath, err := conf.UserUploadPath(s.UserUID, s.RefID+token)

		if err != nil {
			log.Errorf("upload: failed to create storage folder (%s)", clean.Error(err))
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		if err = pruneUploadSidecars(uploadPath); err != nil {
			log.Errorf("upload: could not prepare staged files (%s)", clean.Error(err))
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
		if len(frm.Albums) > 0 && uploadAlbumsAllowed(s) {
			opt.Albums = uploadAlbums(c, s, frm.Albums)
			log.Debugf("upload: adding files to album %s", clean.Log(txt.JoinAnd(opt.Albums)))
		}

		// Set user UID if known.
		if s.UserUID != "" {
			opt.UID = s.UserUID
		}

		// Start import.
		imported := imp.Start(opt)

		// Delete empty import directory.
		if fs.DirIsEmpty(uploadPath) {
			if err = os.Remove(uploadPath); err != nil {
				event.SystemError([]string{"upload", "delete empty folder %s", "%s"}, clean.Log(uploadPath), clean.ErrorFull(err))
			} else {
				event.SystemInfo([]string{"upload", "deleted empty folder %s"}, clean.Log(uploadPath))
			}
		}

		// Update moments if files have been imported.
		if imported.Processed() == 0 {
			event.SystemInfo([]string{"upload", "found no new files in %s"}, clean.Log(uploadPath))
		} else {
			if moments := get.Moments(); moments == nil {
				log.Warnf("upload: moments service not set - you may have found a bug")
			} else if workerErr := moments.Start(); workerErr != nil {
				log.Warnf("moments: %s", clean.Error(workerErr))
			}
		}

		elapsed := time.Since(start)

		log.Infof("library: imported %s in %s", english.Plural(imported.Processed(), "file", "files"), elapsed)

		// Show success message.
		event.PublishSuccessMsg(i18n.MsgUploadProcessed)
		event.PublishCompleted([]string{"import.completed", "index.completed", "upload.completed"}, opt.UID, "", int(elapsed.Seconds()))

		// Update album YAML backups and notify clients of the changes.
		for _, album := range opt.Albums {
			find := entity.AlbumSearch(album, album, entity.AlbumManual)
			find.CreatedBy = opt.UID

			if a := entity.FindAlbum(find); a != nil {
				SaveAlbumYaml(a)
				PublishAlbumEvent(StatusUpdated, a.AlbumUID)
			}
		}

		// Update the user interface.
		UpdateClientConfig()

		// Update album, label, and subject cover thumbs.
		if coversErr := query.UpdateCovers(); coversErr != nil {
			log.Warnf("upload: %s (update covers)", clean.Error(coversErr))
		}

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgUploadProcessed))
	})
}
