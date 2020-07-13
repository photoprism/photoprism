package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
)

// POST /api/v1/photos/:uid/files/:file_uid/unstack
//
// Parameters:
//   uid: string Photo UID as returned by the API
//   file_uid: string File UID as returned by the API
func PhotoFileUnstack(router *gin.RouterGroup) {
	router.POST("/photos/:uid/files/:file_uid/unstack", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		photoUID := c.Param("uid")
		fileUID := c.Param("file_uid")

		file, err := query.FileByUID(fileUID)

		if err != nil {
			log.Errorf("photo: %s (unstack)", err)
			AbortEntityNotFound(c)
			return
		}

		if file.FilePrimary {
			log.Errorf("photo: can't unstack primary file")
			AbortBadRequest(c)
			return
		}

		existingPhoto := *file.Photo
		newPhoto := entity.NewPhoto()

		if err := newPhoto.Create(); err != nil {
			log.Errorf("photo: %s (unstack)", err.Error())
			AbortSaveFailed(c)
			return
		}

		file.Photo = &newPhoto
		file.PhotoID = newPhoto.ID
		file.PhotoUID = newPhoto.PhotoUID

		if err := file.Save(); err != nil {
			log.Errorf("photo: %s (unstack)", err.Error())
			AbortSaveFailed(c)
			return
		}

		fileName := photoprism.FileName(file.FileRoot, file.FileName)

		f, err := photoprism.NewMediaFile(fileName)

		if err != nil {
			log.Errorf("photo: %s (unstack)", err)
			AbortEntityNotFound(c)
			return
		}

		if err := service.Index().MediaFile(f, photoprism.IndexOptions{Rescan: true}, existingPhoto.OriginalName).Error; err != nil {
			log.Errorf("photo: %s (unstack)", err)
			AbortSaveFailed(c)
			return
		}

		PublishPhotoEvent(EntityCreated, file.PhotoUID, c)
		PublishPhotoEvent(EntityUpdated, photoUID, c)

		event.SuccessMsg(i18n.MsgFileUnstacked)

		p, err := query.PhotoPreloadByUID(photoUID)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, p)
	})
}
