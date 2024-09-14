package api

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// DeleteFile removes a file from storage.
//
//	@Summary	removes a file from storage
//	@Id			DeleteFile
//	@Tags		Files
//	@Produce	json
//	@Success	200					{object}	entity.Photo
//	@Failure	401,403,404,429,500	{object}	i18n.Response
//	@Param		uid					path		string	true	"photo uid"
//	@Param		fileuid				path		string	true	"file uid"
//	@Router		/api/v1/photos/{uid}/files/{fileuid} [delete]
func DeleteFile(router *gin.RouterGroup) {
	router.DELETE("/photos/:uid/files/:file_uid", func(c *gin.Context) {
		s := Auth(c, acl.ResourceFiles, acl.ActionDelete)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if conf.ReadOnly() || !conf.Settings().Features.Edit {
			Abort(c, http.StatusForbidden, i18n.ErrReadOnly)
			return
		}

		photoUid := clean.UID(c.Param("uid"))
		fileUid := clean.UID(c.Param("file_uid"))

		file, err := query.FileByUID(fileUid)

		// Found?
		if err != nil {
			log.Errorf("files: %s (delete)", err)
			AbortEntityNotFound(c)
			return
		}

		// Primary file?
		if file.FilePrimary {
			log.Errorf("files: cannot delete primary file")
			AbortDeleteFailed(c)
			return
		}

		// Compose storage filename.
		fileName := photoprism.FileName(file.FileRoot, file.FileName)
		baseName := filepath.Base(fileName)

		mediaFile, err := photoprism.NewMediaFile(fileName)

		if err != nil {
			log.Errorf("files: %s (delete %s)", err, clean.Log(baseName))
			AbortEntityNotFound(c)
			return
		}

		// Report file deletion.
		event.AuditWarn([]string{ClientIP(c), s.UserName, "delete", file.FileName})

		// Remove file from storage.
		if err = mediaFile.Remove(); err != nil {
			log.Errorf("files: %s (delete %s from folder)", err, clean.Log(baseName))
		} else {
			log.Infof("files: deleted %s", clean.Log(baseName))
		}

		// Remove file from index.
		if err = file.Delete(true); err != nil {
			log.Errorf("files: %s (delete %s from index)", err, clean.Log(baseName))
			AbortDeleteFailed(c)
			return
		} else {
			log.Debugf("files: removed %s from index", clean.Log(baseName))
		}

		// Notify clients by publishing events.
		PublishPhotoEvent(StatusUpdated, photoUid, c)

		// Show translated success message.
		event.SuccessMsg(i18n.MsgFileDeleted)

		if p, err := query.PhotoPreloadByUID(photoUid); err != nil {
			AbortEntityNotFound(c)
			return
		} else {
			c.JSON(http.StatusOK, p)
		}
	})
}
