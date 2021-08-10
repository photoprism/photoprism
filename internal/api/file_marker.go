package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/txt"
)

// PUT /api/v1/photos/:uid/files/:file_uid/markers/:id
//
// Parameters:
//   uid: string Photo UID as returned by the API
//   file_uid: string File UID as returned by the API
//   id: int Marker ID as returned by the API
func UpdateFileMarker(router *gin.RouterGroup) {
	router.PUT("/photos/:uid/files/:file_uid/markers/:id", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceFiles, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if !conf.Settings().Features.Edit {
			AbortFeatureDisabled(c)
			return
		}

		photoUID := c.Param("uid")
		fileUID := c.Param("file_uid")
		markerID := txt.UInt(c.Param("id"))

		if photoUID == "" || fileUID == "" || markerID < 1 {
			AbortBadRequest(c)
			return
		}

		file, err := query.FileByUID(fileUID)

		if err != nil {
			log.Errorf("photo: %s (update marker)", err)
			AbortEntityNotFound(c)
			return
		}

		if !file.FilePrimary {
			log.Errorf("photo: can't update markers for non-primary files")
			AbortBadRequest(c)
			return
		} else if file.PhotoUID != photoUID {
			log.Errorf("photo: file uid doesn't match")
			AbortBadRequest(c)
			return
		}

		marker, err := query.MarkerByID(markerID)

		if err != nil {
			log.Errorf("photo: %s (update marker)", err)
			AbortEntityNotFound(c)
			return
		}

		markerForm, err := form.NewMarker(marker)

		if err != nil {
			log.Errorf("photo: %s (new marker form)", err)
			AbortSaveFailed(c)
			return
		}

		if err := c.BindJSON(&markerForm); err != nil {
			log.Errorf("photo: %s (update marker form)", err)
			AbortBadRequest(c)
			return
		}

		if err := marker.SaveForm(markerForm); err != nil {
			log.Errorf("photo: %s (save marker form)", err)
			AbortSaveFailed(c)
			return
		}

		event.SuccessMsg(i18n.MsgChangesSaved)

		if p, err := query.PhotoPreloadByUID(photoUID); err != nil {
			AbortEntityNotFound(c)
			return
		} else {
			if faceCount := file.FaceCount(); p.PhotoFaces == faceCount {
				// Do nothing.
			} else if err := p.Update("PhotoFaces", faceCount); err != nil {
				log.Errorf("photo: %s (update face count)", err)
			} else {
				// Notify clients by publishing events.
				PublishPhotoEvent(EntityUpdated, photoUID, c)

				p.PhotoFaces = faceCount
			}

			c.JSON(http.StatusOK, p)
		}
	})
}
