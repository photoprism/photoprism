package api

import (
	"net/http"
	"path"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/fs"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/pkg/clean"
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

		uid := clean.UID(c.Param("uid"))

		// Users may only change their own avatar.
		if s.User().UserUID != uid {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "user uid does not match"}, s.RefID)
			AbortForbidden(c)
			return
		}

		f, err := c.MultipartForm()

		if err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "%s"}, s.RefID, err)
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		files := f.File["files"]

		if len(files) != 1 {
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		uploadDir, err := conf.UserUploadPath(s.UserUID, "")

		if err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "failed to create folder", "%s"}, s.RefID, err)
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		}

		file := files[0]

		// Uploaded images must be JPEGs with a maximum file size of 20 MB.
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
		} else if !mimeType.Is(fs.MimeTypeJpeg) {
			event.AuditWarn([]string{ClientIP(c), "session %s", "upload avatar", "only jpeg supported"}, s.RefID)
			Abort(c, http.StatusBadRequest, i18n.ErrUnsupportedFormat)
			return
		}

		fileName := "avatar.jpg"
		filePath := path.Join(uploadDir, fileName)

		if err = c.SaveUploadedFile(file, filePath); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "failed to save %s"}, s.RefID, clean.Log(filePath))
			Abort(c, http.StatusBadRequest, i18n.ErrUploadFailed)
			return
		} else {
			event.AuditInfo([]string{ClientIP(c), "session %s", "upload avatar", "saved as %s"}, s.RefID, clean.Log(filePath))
		}

		if mediaFile, mediaErr := photoprism.NewMediaFile(filePath); mediaErr != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "%s"}, s.RefID, err)
			Abort(c, http.StatusBadRequest, i18n.ErrUnsupportedFormat)
			return
		} else if err = mediaFile.CreateThumbnails(conf.ThumbCachePath(), false); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "%s"}, s.RefID, err)
		} else if err = s.User().SetAvatar(mediaFile.Hash(), entity.SrcManual); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "upload avatar", "%s"}, s.RefID, err)
		}

		// Clear the session cache, as it contains user information.
		s.ClearCache()

		log.Info(i18n.Msg(i18n.MsgFileUploaded))

		c.JSON(http.StatusOK, entity.FindUserByUID(uid))
	})
}
