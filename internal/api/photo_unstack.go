package api

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/txt"
)

// POST /api/v1/photos/:uid/files/:file_uid/unstack
//
// Parameters:
//   uid: string Photo UID as returned by the API
//   file_uid: string File UID as returned by the API
func PhotoUnstack(router *gin.RouterGroup) {
	router.POST("/photos/:uid/files/:file_uid/unstack", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

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
		} else if file.FileSidecar {
			log.Errorf("photo: can't unstack sidecar files")
			AbortBadRequest(c)
			return
		} else if file.FileRoot != entity.RootOriginals {
			log.Errorf("photo: only originals can be unstacked")
			AbortBadRequest(c)
			return
		}

		fileName := photoprism.FileName(file.FileRoot, file.FileName)
		baseName := filepath.Base(fileName)

		mediaFile, err := photoprism.NewMediaFile(fileName)

		if err != nil {
			log.Errorf("photo: %s (unstack %s)", err, txt.Quote(baseName))
			AbortEntityNotFound(c)
			return
		}

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			log.Errorf("photo: %s (unstack %s)", err, txt.Quote(baseName))
			AbortEntityNotFound(c)
			return
		} else if related.Len() == 0 {
			log.Errorf("photo: no related files found (unstack %s)", txt.Quote(baseName))
			AbortEntityNotFound(c)
			return
		}

		if related.Len() > 1 {
			if conf.ReadOnly() {
				log.Errorf("photo: can't rename files in read only mode (unstack %s)", txt.Quote(baseName))
				AbortFeatureDisabled(c)
				return
			}

			newName := mediaFile.AbsPrefix(true) + "_" + mediaFile.Checksum() + mediaFile.Extension()

			if err := mediaFile.Move(newName); err != nil {
				log.Errorf("photo: can't rename %s to %s (unstack)", txt.Quote(baseName), txt.Quote(filepath.Base(newName)))
				AbortUnexpected(c)
				return
			}
		}

		oldPhoto := *file.Photo
		oldPrimary, err := oldPhoto.PrimaryFile()

		if err != nil {
			log.Errorf("photo: can't find primary file for existing photo (unstack %s)", txt.Quote(baseName))
			AbortUnexpected(c)
			return
		}

		newPhoto := entity.NewPhoto()

		if err := newPhoto.Create(); err != nil {
			log.Errorf("photo: %s (unstack %s)", err.Error(), txt.Quote(baseName))
			AbortSaveFailed(c)
			return
		}

		file.Photo = &newPhoto
		file.PhotoID = newPhoto.ID
		file.PhotoUID = newPhoto.PhotoUID
		file.FileName = mediaFile.RelName(conf.OriginalsPath())

		if err := file.Save(); err != nil {
			log.Errorf("photo: %s (unstack %s)", err.Error(), txt.Quote(baseName))

			if err := newPhoto.Delete(true); err != nil {
				log.Errorf("photo: %s (unstack %s)", err.Error(), txt.Quote(baseName))
			}

			AbortSaveFailed(c)
			return
		}

		ind := service.Index()

		// Index new, unstacked file.
		if res := ind.File(mediaFile.FileName()); res.Failed() {
			log.Errorf("photo: %s (unstack %s)", res.Err, txt.Quote(baseName))
			AbortSaveFailed(c)
			return
		}

		// Reset type for old, existing photo to image.
		if err := oldPhoto.Update("PhotoType", entity.TypeImage); err != nil {
			log.Errorf("photo: %s (unstack %s)", err, txt.Quote(baseName))
			AbortUnexpected(c)
			return
		}

		// Get name of old, existing primary file.
		oldPrimaryName := photoprism.FileName(oldPrimary.FileRoot, oldPrimary.FileName)

		// Re-index old, existing primary file.
		if res := ind.File(oldPrimaryName); res.Failed() {
			log.Errorf("photo: %s (unstack %s)", res.Err, txt.Quote(baseName))
			AbortSaveFailed(c)
			return
		}

		// Notify clients by publishing events.
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
