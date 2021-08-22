package api

import (
	"net/http"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/txt"
)

// findFileMarker returns a file and marker entity matching the api request.
func findFileMarker(c *gin.Context) (file entity.File, marker entity.Marker, err error) {
	s := Auth(SessionID(c), acl.ResourceFiles, acl.ActionUpdate)

	if s.Invalid() {
		AbortUnauthorized(c)
		return file, marker, err
	}

	conf := service.Config()

	if !conf.Settings().Features.Edit {
		AbortFeatureDisabled(c)
		return file, marker, err
	}

	photoUID := c.Param("uid")
	fileUID := c.Param("file_uid")
	markerID := txt.UInt(c.Param("id"))

	if photoUID == "" || fileUID == "" || markerID < 1 {
		AbortBadRequest(c)
		return file, marker, err
	}

	file, err = query.FileByUID(fileUID)

	if err != nil {
		log.Errorf("photo: %s (update marker)", err)
		AbortEntityNotFound(c)
		return file, marker, err
	}

	if !file.FilePrimary {
		log.Errorf("photo: can't update markers for non-primary files")
		AbortBadRequest(c)
		return file, marker, err
	} else if file.PhotoUID != photoUID {
		log.Errorf("photo: file uid doesn't match")
		AbortBadRequest(c)
		return file, marker, err
	}

	marker, err = query.MarkerByID(markerID)

	if err != nil {
		log.Errorf("photo: %s (update marker)", err)
		AbortEntityNotFound(c)
		return file, marker, err
	}

	return file, marker, nil
}

// UpdateMarker updates an existing file marker e.g. representing a face.
//
// PUT /api/v1/photos/:uid/files/:file_uid/markers/:id
//
// Parameters:
//   uid: string Photo UID as returned by the API
//   file_uid: string File UID as returned by the API
//   id: int Marker ID as returned by the API
func UpdateMarker(router *gin.RouterGroup) {
	router.PUT("/photos/:uid/files/:file_uid/markers/:id", func(c *gin.Context) {
		file, marker, err := findFileMarker(c)

		if err != nil {
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
		} else if marker.SubjectUID != "" && marker.SubjectSrc == entity.SrcManual && marker.FaceID != "" {
			if res, err := service.Faces().Optimize(); err != nil {
				log.Errorf("faces: %s (optimize)", err)
			} else if res.Merged > 0 {
				log.Infof("faces: %d clusters merged", res.Merged)
			}
		}

		event.SuccessMsg(i18n.MsgChangesSaved)

		if p, err := query.PhotoPreloadByUID(file.PhotoUID); err != nil {
			AbortEntityNotFound(c)
			return
		} else {
			if faceCount := file.FaceCount(); p.PhotoFaces == faceCount {
				// Do nothing.
			} else if err := p.Update("PhotoFaces", faceCount); err != nil {
				log.Errorf("photo: %s (update face count)", err)
			} else {
				// Notify clients by publishing events.
				PublishPhotoEvent(EntityUpdated, file.PhotoUID, c)

				p.PhotoFaces = faceCount
			}

			c.JSON(http.StatusOK, p)
		}
	})
}

// ClearMarkerSubject removes an existing marker subject association.
//
// DELETE /api/v1/photos/:uid/files/:file_uid/markers/:id/subject
//
// Parameters:
//   uid: string Photo UID as returned by the API
//   file_uid: string File UID as returned by the API
//   id: int Marker ID as returned by the API
func ClearMarkerSubject(router *gin.RouterGroup) {
	router.DELETE("/photos/:uid/files/:file_uid/markers/:id/subject", func(c *gin.Context) {
		_, marker, err := findFileMarker(c)

		if err != nil {
			return
		}

		if err := marker.ClearSubject(entity.SrcManual); err != nil {
			log.Errorf("faces: %s (clear subject)", err)
			AbortSaveFailed(c)
			return
		}

		event.SuccessMsg(i18n.MsgChangesSaved)

		c.JSON(http.StatusOK, marker)
	})
}
