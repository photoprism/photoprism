package api

import (
	"net/http"
	"path"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// UploadUserAvatar updates the avatar image of the currently authenticated user.
//
// POST /api/v1/users/:uid/avatar
func UploadUserAvatar(router *gin.RouterGroup) {
	router.POST("/users/:uid/avatar", func(c *gin.Context) {
		conf := get.Config()

		if conf.Demo() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		s := AuthAny(c, acl.ResourceUsers, acl.Permissions{acl.ActionManage, acl.AccessOwn})

		if s.Abort(c) {
			return
		}

		// Check if the session user is has user management privileges.
		isPrivileged := acl.Resources.AllowAll(acl.ResourceUsers, s.User().AclRole(), acl.Permissions{acl.AccessAll, acl.ActionManage})
		uid := clean.UID(c.Param("uid"))

		// Users may only change their own avatar.
		if !isPrivileged && s.User().UserUID != uid {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "user does not match"}, s.RefID)
			AbortForbidden(c)
			return
		}

		// Parse upload form.
		f, err := c.MultipartForm()

		if err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "%s"}, s.RefID, err)
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		// Check number of files.
		files := f.File["files"]

		if len(files) != 1 {
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		// Find user entity to update.
		m := entity.FindUserByUID(uid)

		if m == nil {
			Abort(c, http.StatusNotFound, i18n.ErrUserNotFound)
			return
		}

		// Get user upload folder.
		uploadDir, err := conf.UserUploadPath(uid, "")

		if err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "failed to create folder", "%s"}, s.RefID, err)
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		file := files[0]
		var fileName string

		// The user avatar must be a PNG or JPEG image with a maximum size of 20 MB.
		if file.Size > 20000000 {
			event.AuditWarn([]string{ClientIP(c), "session %s", "upload avatar", "file size exceeded"}, s.RefID)
			Abort(c, http.StatusBadRequest, i18n.ErrFileTooLarge)
			return
		} else if fReader, fErr := file.Open(); fErr != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "%s"}, s.RefID, err)
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		} else if mimeType, mimeErr := mimetype.DetectReader(fReader); mimeErr != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "%s"}, s.RefID, err)
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		} else {
			switch {
			case mimeType.Is(fs.MimeTypePNG):
				fileName = "avatar.png"
			case mimeType.Is(fs.MimeTypeJPEG):
				fileName = "avatar.jpg"
			default:
				event.AuditWarn([]string{ClientIP(c), "session %s", "upload avatar", " %s not supported"}, s.RefID, mimeType)
				Abort(c, http.StatusBadRequest, i18n.ErrUnsupportedFormat)
				return
			}
		}

		// Get absolute file path.
		filePath := path.Join(uploadDir, fileName)

		// Save avatar image.
		if err = c.SaveUploadedFile(file, filePath); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "failed to save %s"}, s.RefID, clean.Log(filePath))
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		} else {
			event.AuditInfo([]string{ClientIP(c), "session %s", "upload avatar", "saved as %s"}, s.RefID, clean.Log(filePath))
		}

		// Create avatar thumbnails.
		if mediaFile, mediaErr := photoprism.NewMediaFile(filePath); mediaErr != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "%s"}, s.RefID, err)
			Abort(c, http.StatusBadRequest, i18n.ErrUnsupportedFormat)
			return
		} else if err = mediaFile.CreateThumbnails(conf.ThumbCachePath(), false); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "%s"}, s.RefID, err)
		} else if err = m.SetAvatar(mediaFile.Hash(), entity.SrcManual); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "%s"}, s.RefID, err)
		}

		// Clear session cache to update user details.
		s.ClearCache()

		// Show success message.
		log.Info(i18n.Msg(i18n.MsgFileUploaded))

		// Return updated user profile.
		c.JSON(http.StatusOK, entity.FindUserByUID(uid))
	})
}
