package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
)

// Checks if background worker runs less than once per hour.
func wakeupIntervalTooHigh(c *gin.Context) bool {
	if conf := get.Config(); conf.Unsafe() {
		return false
	} else if i := conf.WakeupInterval(); i > time.Hour {
		Abort(c, http.StatusForbidden, i18n.ErrWakeupInterval, i.String())
		return true
	} else {
		return false
	}
}

// findFileMarker returns a file and marker entity matching the api request.
func findFileMarker(c *gin.Context) (file *entity.File, marker *entity.Marker, err error) {
	// Check authorization.
	s := Auth(c, acl.ResourceFiles, acl.ActionUpdate)

	if s.Invalid() {
		AbortForbidden(c)
		return nil, nil, fmt.Errorf("unauthorized")
	}

	// Check feature flags.
	conf := get.Config()
	if !conf.Settings().Features.People {
		AbortFeatureDisabled(c)
		return nil, nil, fmt.Errorf("feature disabled")
	}

	// Find marker.
	if uid := c.Param("marker_uid"); uid == "" {
		AbortBadRequest(c)
		return nil, nil, fmt.Errorf("bad request")
	} else if marker, err = query.MarkerByUID(uid); err != nil {
		AbortEntityNotFound(c)
		return nil, nil, fmt.Errorf("uid %s %s", uid, err)
	} else if marker.FileUID == "" {
		AbortEntityNotFound(c)
		return nil, marker, fmt.Errorf("marker file missing")
	}

	// Find file.
	if file, err = query.FileByUID(marker.FileUID); err != nil {
		AbortEntityNotFound(c)
		return file, marker, fmt.Errorf("file %s %s", marker.FileUID, err)
	}

	return file, marker, nil
}

// UpdateMarker updates an existing file marker e.g. representing a face.
//
// PUT /api/v1/markers/:marker_uid
//
// Parameters:
//
//	uid: string Photo UID as returned by the API
//	file_uid: string File UID as returned by the API
//	id: int Marker ID as returned by the API
func UpdateMarker(router *gin.RouterGroup) {
	router.PUT("/markers/:marker_uid", func(c *gin.Context) {
		// Abort if workers runs less than once per hour.
		if wakeupIntervalTooHigh(c) {
			return
		}

		// Abort if another update is running.
		if err := mutex.UpdatePeople.Start(); err != nil {
			AbortBusy(c)
			return
		}

		defer mutex.UpdatePeople.Stop()

		file, marker, err := findFileMarker(c)

		if err != nil {
			log.Debugf("faces: %s (find marker to update)", err)
			return
		}

		// Initialize form.
		f, err := form.NewMarker(*marker)

		if err != nil {
			log.Errorf("faces: %s (create marker update form)", err)
			AbortSaveFailed(c)
			return
		} else if err := c.BindJSON(&f); err != nil {
			log.Errorf("faces: %s (set updated marker values)", err)
			AbortBadRequest(c)
			return
		}

		// Update marker from form values.
		if changed, err := marker.SaveForm(f); err != nil {
			log.Errorf("faces: %s (update marker)", err)
			AbortSaveFailed(c)
			return
		} else if changed {
			if marker.FaceID != "" && marker.SubjUID != "" && marker.SubjSrc == entity.SrcManual {
				if res, err := get.Faces().Optimize(); err != nil {
					log.Errorf("faces: %s (optimize)", err)
				} else if res.Merged > 0 {
					log.Infof("faces: merged %s", english.Plural(res.Merged, "cluster", "clusters"))
				}
			}

			if err := query.UpdateSubjectCovers(); err != nil {
				log.Errorf("faces: %s (update covers)", err)
			}

			if err := entity.UpdateSubjectCounts(); err != nil {
				log.Errorf("faces: %s (update counts)", err)
			}
		}

		// Update photo metadata.
		if !file.FilePrimary {
			log.Infof("faces: skipped updating photo for non-primary file")
		} else if p, err := query.PhotoByUID(file.PhotoUID); err != nil {
			log.Errorf("faces: %s (find photo))", err)
		} else if err := p.UpdateAndSaveTitle(); err != nil {
			log.Errorf("faces: %s (update photo title)", err)
		} else {
			// Notify clients.
			PublishPhotoEvent(EntityUpdated, file.PhotoUID, c)
		}

		event.SuccessMsg(i18n.MsgChangesSaved)

		c.JSON(http.StatusOK, marker)
	})
}

// ClearMarkerSubject removes an existing marker subject association.
//
// DELETE /api/v1/markers/:marker_uid/subject
//
// Parameters:
//
//	uid: string Photo UID as returned by the API
//	file_uid: string File UID as returned by the API
//	id: int Marker ID as returned by the API
func ClearMarkerSubject(router *gin.RouterGroup) {
	router.DELETE("/markers/:marker_uid/subject", func(c *gin.Context) {
		// Abort if workers runs less than once per hour.
		if wakeupIntervalTooHigh(c) {
			return
		}

		// Abort if another update is running.
		if err := mutex.UpdatePeople.Start(); err != nil {
			AbortBusy(c)
			return
		}

		defer mutex.UpdatePeople.Stop()

		file, marker, err := findFileMarker(c)

		if err != nil {
			log.Debugf("faces: %s (find marker to clear subject)", err)
			return
		}

		if err := marker.ClearSubject(entity.SrcManual); err != nil {
			log.Errorf("faces: %s (clear marker subject)", err)
			AbortSaveFailed(c)
			return
		} else if err := query.UpdateSubjectCovers(); err != nil {
			log.Errorf("faces: %s (update covers)", err)
		} else if err := entity.UpdateSubjectCounts(); err != nil {
			log.Errorf("faces: %s (update counts)", err)
		}

		// Update photo metadata.
		if !file.FilePrimary {
			log.Infof("faces: skipped updating photo for non-primary file")
		} else if p, err := query.PhotoByUID(file.PhotoUID); err != nil {
			log.Errorf("faces: %s (find photo))", err)
		} else if err := p.UpdateAndSaveTitle(); err != nil {
			log.Errorf("faces: %s (update photo title)", err)
		} else {
			// Notify clients.
			PublishPhotoEvent(EntityUpdated, file.PhotoUID, c)
		}

		event.SuccessMsg(i18n.MsgChangesSaved)

		c.JSON(http.StatusOK, marker)
	})
}
